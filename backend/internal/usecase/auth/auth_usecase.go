package auth

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"
	"anirank/api/internal/infrastructure/anilist"
	"anirank/api/internal/infrastructure/discord"
	"anirank/api/internal/infrastructure/google"
	"anirank/api/internal/infrastructure/security"
	"anirank/api/internal/pkg/avatar"
	"anirank/api/internal/pkg/crypto"
	"anirank/api/internal/pkg/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase struct {
	userRepo      domain.UserRepository
	jwtService    *JWTService
	storage       infrastructure.StorageService
	media         infrastructure.MediaService
	xpUsecase     domain.XPUsecase
	badgeUsecase  domain.BadgeUsecase
	anilist       anilist.AnilistClient
	google        google.GoogleClient
	discord       discord.DiscordClient
	encryptionKey string
}

func NewAuthUsecase(userRepo domain.UserRepository, jwtService *JWTService, storage infrastructure.StorageService, media infrastructure.MediaService, xu domain.XPUsecase, bu domain.BadgeUsecase, ac anilist.AnilistClient, gc google.GoogleClient, dc discord.DiscordClient, eKey string) *AuthUsecase {
	return &AuthUsecase{
		userRepo:      userRepo,
		jwtService:    jwtService,
		storage:       storage,
		media:         media,
		xpUsecase:     xu,
		badgeUsecase:  bu,
		anilist:       ac,
		google:        gc,
		discord:       dc,
		encryptionKey: eKey,
	}
}

// AuthTokenResponse structures the payload delivered to client upon successful auth
type AuthTokenResponse struct {
	Token string       `json:"token"`
	User  *domain.User `json:"user"`
}

func (u *AuthUsecase) Login(ctx context.Context, email, password string) (*AuthTokenResponse, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, domain.NewAppError(401, "Invalid email or password", nil)
		}
		return nil, domain.NewAppError(500, "Database error", err)
	}

	// Verify Hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, domain.NewAppError(401, "Invalid email or password", nil)
	}

	roleSlugs := []string{}
	roles, _ := u.userRepo.GetRolesByUserID(ctx, user.ID)
	for _, r := range roles {
		roleSlugs = append(roleSlugs, r.Slug)
	}
	if len(roleSlugs) == 0 {
		roleSlugs = append(roleSlugs, "user") // Fallback
	}
	user.Roles = roles

	// Check daily login status
	_ = u.xpUsecase.CheckDailyLogin(ctx, user.ID)

	// Fetch user again to get updated XP/Level if it was awarded
	if refreshed, err := u.userRepo.GetByID(ctx, user.ID); err == nil {
		refreshed.Roles = roles // Re-attach roles after refresh
		user = refreshed
	}

	// Generate standard token
	token, err := u.jwtService.GenerateToken(user.UUID, roleSlugs)
	if err != nil {
		return nil, domain.NewAppError(500, "Could not generate authentication token", err)
	}

	// Scrub password from response
	user.Password = ""

	// Default score format for users who haven't set one
	if user.ScoreFormat == nil || *user.ScoreFormat == "" {
		defaultFormat := "POINT_10_DECIMAL"
		user.ScoreFormat = &defaultFormat
	}

	u.enrichUserImages(user)

	return &AuthTokenResponse{
		Token: token,
		User:  user,
	}, nil
}

func (u *AuthUsecase) Register(ctx context.Context, name, email, password string) (*AuthTokenResponse, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	// Ensure email is unique
	existing, _ := u.userRepo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, domain.NewAppError(400, "Email is already registered", nil)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, domain.NewAppError(500, "Failed to secure password", err)
	}

	sFormat := "POINT_10_DECIMAL"

	newUser := &domain.User{
		UUID:        uuid.New().String(),
		Name:        name,
		Email:       email,
		Password:    string(hash),
		ScoreFormat: &sFormat,
	}

	slug := u.generateUniqueUserSlug(ctx, name)
	newUser.Slug = &slug

	if err := u.userRepo.Create(ctx, newUser); err != nil {
		return nil, domain.NewAppError(500, "Failed to create user account", err)
	}

	token, err := u.jwtService.GenerateToken(newUser.UUID, []string{"user"}) // New users are strictly basic tier
	if err != nil {
		return nil, domain.NewAppError(500, "Could not generate authentication token", err)
	}

	newUser.Password = ""

	// Generate avatar in background
	go u.GenerateUserAvatar(context.Background(), newUser.ID, newUser.Name)

	return &AuthTokenResponse{
		Token: token,
		User:  newUser,
	}, nil
}

func (u *AuthUsecase) GenerateUserAvatar(ctx context.Context, userID uint64, name string) {
	res, err := avatar.Generate(ctx, name, 256)
	if err != nil {
		log.Printf("ERROR: failed to generate avatar for user %d: %v", userID, err)
		return
	}

	buf := bytes.NewReader(res.Data)
	filename := fmt.Sprintf("users/avatars/%d_%s.avif", userID, uuid.New().String())

	_, err = u.storage.UploadFile(ctx, filename, buf, res.Size, res.ContentType)
	if err != nil {
		log.Printf("ERROR: failed to upload generated avatar for user %d: %v", userID, err)
		return
	}

	if err := u.userRepo.SetImage(ctx, userID, "avatar", filename); err != nil {
		log.Printf("ERROR: failed to save user avatar reference in DB for user %d: %v", userID, err)
	}
}

func (u *AuthUsecase) GetProfile(ctx context.Context, userID uint64) (*domain.User, error) {
	// Check daily login XP on profile retrieval (backup for persistent sessions)
	_ = u.xpUsecase.CheckDailyLogin(ctx, userID)

	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, domain.NewAppError(404, "User profile not found", err)
	}

	user.Password = "" // NEVER retrieve password hashes to JSON clients

	// Default score format for users who haven't set one
	if user.ScoreFormat == nil || *user.ScoreFormat == "" {
		defaultFormat := "POINT_10_DECIMAL"
		user.ScoreFormat = &defaultFormat
	}

	// Load Roles for RBAC hydration
	roles, _ := u.userRepo.GetRolesByUserID(ctx, user.ID)
	user.Roles = roles

	u.enrichUserImages(user)

	return user, nil
}

func (u *AuthUsecase) enrichUserImages(user *domain.User) {
	if user == nil {
		return
	}
	user.AvatarUrl = u.media.Resolve(user.Avatar)
	user.BannerUrl = u.media.Resolve(user.Banner)

	if user.Avatar != nil {
		user.AvatarSources = u.media.GetImageSources(*user.Avatar)
	}
	if user.Banner != nil {
		user.BannerSources = u.media.GetImageSources(*user.Banner)
	}
}

func (u *AuthUsecase) UpdateAvatar(ctx context.Context, userID uint64, file io.Reader, size int64, contentType string) (string, error) {
	if u.storage == nil {
		return "", domain.NewAppError(500, "Storage service is not configured", nil)
	}

	// Fetch user to get current avatar for cleanup
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", domain.NewAppError(404, "User not found", err)
	}

	path, url, err := u.media.UploadWithResolutions(ctx, "users/avatars", userID, file, infrastructure.PresetSquare)
	if err != nil {
		return "", domain.NewAppError(500, "Failed to upload avatar", err)
	}

	// Delete old avatar if it exists
	if user.Avatar != nil {
		u.media.DeleteMedia(ctx, *user.Avatar)
	}

	if err := u.userRepo.SetImage(ctx, userID, "avatar", path); err != nil {
		return "", domain.NewAppError(500, "Failed to save avatar reference", err)
	}

	return url, nil
}

func (u *AuthUsecase) UpdateBanner(ctx context.Context, userID uint64, file io.Reader, size int64, contentType string) (string, error) {
	if u.storage == nil {
		return "", domain.NewAppError(500, "Storage service is not configured", nil)
	}

	// Fetch user to get current banner for cleanup
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", domain.NewAppError(404, "User not found", err)
	}

	path, url, err := u.media.UploadWithResolutions(ctx, "users/banners", userID, file, infrastructure.PresetLandscape)
	if err != nil {
		return "", domain.NewAppError(500, "Failed to upload banner", err)
	}

	// Delete old banner if it exists
	if user.Banner != nil {
		u.media.DeleteMedia(ctx, *user.Banner)
	}

	if err := u.userRepo.SetImage(ctx, userID, "banner", path); err != nil {
		return "", domain.NewAppError(500, "Failed to save banner reference", err)
	}

	return url, nil
}

func (u *AuthUsecase) UpdateScoreFormat(ctx context.Context, userID uint64, format string) error {
	return u.userRepo.UpdateScoreFormat(ctx, userID, format)
}

func (u *AuthUsecase) UpdateProfile(ctx context.Context, userID uint64, about, profileColor *string) error {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if about != nil {
		user.About = security.SanitizeStrictPtr(about)
	}
	if profileColor != nil {
		user.ProfileColor = profileColor
	}

	return u.userRepo.Update(ctx, user)
}

func (u *AuthUsecase) generateUniqueUserSlug(ctx context.Context, name string) string {
	return utils.GenerateUniqueSlug(name, func(slug string) bool {
		existing, err := u.userRepo.GetBySlug(ctx, slug)
		return err == nil && existing != nil
	})
}

// anilistRedirectURI returns OAuth redirect: ANILIST_REDIRECT_URL, or APP_URL+"/settings/account".
// Must match exactly what is registered at https://anilist.co/settings/developer
// and the page that handles ?code= (e.g. settings/account/+page.svelte).
func anilistRedirectURI() string {
	if u := strings.TrimSpace(os.Getenv("ANILIST_REDIRECT_URL")); u != "" {
		return u
	}
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("APP_URL")), "/")
	if base == "" {
		return ""
	}
	return base + "/settings/account"
}

// AnilistAuthURL builds the AniList OAuth2 authorize URL for "link account".
func (u *AuthUsecase) AnilistAuthURL() (string, error) {
	clientID := strings.TrimSpace(os.Getenv("ANILIST_CLIENT_ID"))
	if clientID == "" {
		return "", fmt.Errorf("ANILIST_CLIENT_ID is not set")
	}
	redirectURI := anilistRedirectURI()
	if redirectURI == "" {
		return "", fmt.Errorf("set ANILIST_REDIRECT_URL or APP_URL for OAuth redirect (e.g. https://yoursite.com/settings/account)")
	}
	authURL := fmt.Sprintf(
		"https://anilist.co/api/v2/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&state=%s",
		url.QueryEscape(clientID),
		url.QueryEscape(redirectURI),
		url.QueryEscape("anilistrank_link"),
	)
	return authURL, nil
}

// anilistLoginRedirectURI is the OAuth redirect for sign-in (must be registered in AniList developer app).
func anilistLoginRedirectURI() string {
	if u := strings.TrimSpace(os.Getenv("ANILIST_LOGIN_REDIRECT_URL")); u != "" {
		return u
	}
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("APP_URL")), "/")
	if base == "" {
		return ""
	}
	return base + "/login"
}

// AnilistLoginAuthURL builds the AniList OAuth2 authorize URL for "Login with AniList".
func (u *AuthUsecase) AnilistLoginAuthURL() (string, error) {
	clientID := strings.TrimSpace(os.Getenv("ANILIST_CLIENT_ID"))
	if clientID == "" {
		return "", fmt.Errorf("ANILIST_CLIENT_ID is not set")
	}
	redirectURI := anilistLoginRedirectURI()
	if redirectURI == "" {
		return "", fmt.Errorf("set ANILIST_LOGIN_REDIRECT_URL or APP_URL for login redirect (e.g. https://yoursite.com/login)")
	}
	authURL := fmt.Sprintf(
		"https://anilist.co/api/v2/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&state=%s",
		url.QueryEscape(clientID),
		url.QueryEscape(redirectURI),
		url.QueryEscape("anilistrank_login"),
	)
	return authURL, nil
}

func (u *AuthUsecase) LinkAnilist(ctx context.Context, userID uint64, code string) error {
	clientID := strings.TrimSpace(os.Getenv("ANILIST_CLIENT_ID"))
	clientSecret := strings.TrimSpace(os.Getenv("ANILIST_CLIENT_SECRET"))
	redirectURI := anilistRedirectURI()

	if clientID == "" || clientSecret == "" {
		return domain.NewAppError(http.StatusInternalServerError, "AniList OAuth is not configured (missing ANILIST_CLIENT_ID or ANILIST_CLIENT_SECRET)", nil)
	}
	if redirectURI == "" {
		return domain.NewAppError(http.StatusInternalServerError, "AniList redirect URI not configured (set ANILIST_REDIRECT_URL or APP_URL)", nil)
	}

	// 1. Exchange code for tokens
	tokenResp, err := u.anilist.ExchangeCode(ctx, clientID, clientSecret, redirectURI, strings.TrimSpace(code))
	if err != nil {
		return domain.NewAppError(http.StatusBadRequest, "Failed to exchange Anilist code", err)
	}

	// 2. Get user info from Anilist
	anilistUser, err := u.anilist.GetViewer(ctx, tokenResp.AccessToken)
	if err != nil {
		return domain.NewAppError(http.StatusInternalServerError, "Failed to fetch Anilist profile", err)
	}

	// 3. Encrypt tokens
	encryptedAccess, err := crypto.Encrypt(tokenResp.AccessToken, u.encryptionKey)
	if err != nil {
		return domain.NewAppError(http.StatusInternalServerError, "Failed to encrypt access token", err)
	}

	encryptedRefresh := ""
	if tokenResp.RefreshToken != "" {
		encryptedRefresh, err = crypto.Encrypt(tokenResp.RefreshToken, u.encryptionKey)
		if err != nil {
			return domain.NewAppError(http.StatusInternalServerError, "Failed to encrypt refresh token", err)
		}
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	// 4. Save social identity
	identity := &domain.UserSocialIdentity{
		UserID:           userID,
		Provider:         "anilist",
		ProviderID:       fmt.Sprintf("%d", anilistUser.ID),
		ProviderUsername: &anilistUser.Name,
		AccessToken:      &encryptedAccess,
		RefreshToken:     &encryptedRefresh,
		ExpiresAt:        &expiresAt,
	}
	if err := u.userRepo.SaveSocialIdentity(ctx, identity); err != nil {
		return err
	}

	// Automatic Badge Check
	_ = u.badgeUsecase.CheckAndAwardBadges(ctx, userID, "anilist")

	return nil
}

func (u *AuthUsecase) LinkGoogle(ctx context.Context, userID uint64, code string) error {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURI := os.Getenv("GOOGLE_REDIRECT_URL")

	// 1. Exchange code for tokens
	tokenResp, err := u.google.ExchangeCode(ctx, clientID, clientSecret, redirectURI, code)
	if err != nil {
		return domain.NewAppError(http.StatusBadRequest, "Failed to exchange Google code", err)
	}

	// 2. Get user info from Google
	googleUser, err := u.google.GetUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		return domain.NewAppError(http.StatusInternalServerError, "Failed to fetch Google profile", err)
	}

	// 3. Encrypt tokens
	encryptedAccess, err := crypto.Encrypt(tokenResp.AccessToken, u.encryptionKey)
	if err != nil {
		return domain.NewAppError(http.StatusInternalServerError, "Failed to encrypt access token", err)
	}

	encryptedRefresh := ""
	if tokenResp.RefreshToken != "" {
		encryptedRefresh, err = crypto.Encrypt(tokenResp.RefreshToken, u.encryptionKey)
		if err != nil {
			return domain.NewAppError(http.StatusInternalServerError, "Failed to encrypt refresh token", err)
		}
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	// 4. Save social identity
	identity := &domain.UserSocialIdentity{
		UserID:           userID,
		Provider:         "google",
		ProviderID:       googleUser.Sub,
		ProviderUsername: &googleUser.Email,
		AccessToken:      &encryptedAccess,
		RefreshToken:     &encryptedRefresh,
		ExpiresAt:        &expiresAt,
	}
	return u.userRepo.SaveSocialIdentity(ctx, identity)
}

func (u *AuthUsecase) LoginWithGoogle(ctx context.Context, code, redirectURI string) (*AuthTokenResponse, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	// 1. Exchange code for tokens
	tokenResp, err := u.google.ExchangeCode(ctx, clientID, clientSecret, redirectURI, code)
	if err != nil {
		return nil, domain.NewAppError(http.StatusBadRequest, "Failed to exchange Google code", err)
	}

	// 2. Get user info from Google
	googleUser, err := u.google.GetUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, domain.NewAppError(http.StatusInternalServerError, "Failed to fetch Google profile", err)
	}

	// 3. Find user by google_id
	user, err := u.userRepo.GetByGoogleID(ctx, googleUser.Sub)
	if err != nil {
		if err == domain.ErrNotFound {
			// 4. Fallback: find by email (if verified)
			if googleUser.EmailVerified {
				user, err = u.userRepo.GetByEmail(ctx, googleUser.Email)
				if err != nil {
					if err == domain.ErrNotFound {
						// 4. Auto-register if not found by Google ID or Email
						return u.autoRegisterGoogleUser(ctx, googleUser, tokenResp)
					}
					return nil, domain.NewAppError(http.StatusInternalServerError, "Database error", err)
				}
				
				// Automatically link the Google ID to this user using the new table
				encryptedAccess, _ := crypto.Encrypt(tokenResp.AccessToken, u.encryptionKey)
				encryptedRefresh := ""
				if tokenResp.RefreshToken != "" {
					encryptedRefresh, _ = crypto.Encrypt(tokenResp.RefreshToken, u.encryptionKey)
				}
				expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

				identity := &domain.UserSocialIdentity{
					UserID:           user.ID,
					Provider:         "google",
					ProviderID:       googleUser.Sub,
					ProviderUsername: &googleUser.Email,
					AccessToken:      &encryptedAccess,
					RefreshToken:     &encryptedRefresh,
					ExpiresAt:        &expiresAt,
				}

				if err := u.userRepo.SaveSocialIdentity(ctx, identity); err != nil {
					log.Printf("Failed to update user google info during login: %v", err)
				}
			} else {
				return nil, domain.NewAppError(http.StatusUnauthorized, "Google account not linked and email not verified", nil)
			}
		} else {
			return nil, domain.NewAppError(http.StatusInternalServerError, "Database error", err)
		}
	}

	// 5. Generate JWT
	roleSlugs := []string{}
	roles, _ := u.userRepo.GetRolesByUserID(ctx, user.ID)
	for _, r := range roles {
		roleSlugs = append(roleSlugs, r.Slug)
	}
	if len(roleSlugs) == 0 {
		roleSlugs = append(roleSlugs, "user") // Fallback
	}
	user.Roles = roles

	token, err := u.jwtService.GenerateToken(user.UUID, roleSlugs)
	if err != nil {
		return nil, domain.NewAppError(http.StatusInternalServerError, "Failed to generate token", err)
	}

	user.Password = ""
	u.enrichUserImages(user)

	return &AuthTokenResponse{
		Token: token,
		User:  user,
	}, nil
}

// LoginWithAnilist completes OAuth from /login (redirect URI must match AnilistLoginAuthURL).
func (u *AuthUsecase) LoginWithAnilist(ctx context.Context, code string) (*AuthTokenResponse, error) {
	clientID := strings.TrimSpace(os.Getenv("ANILIST_CLIENT_ID"))
	clientSecret := strings.TrimSpace(os.Getenv("ANILIST_CLIENT_SECRET"))
	redirectURI := anilistLoginRedirectURI()

	if clientID == "" || clientSecret == "" {
		return nil, domain.NewAppError(http.StatusInternalServerError, "AniList OAuth is not configured", nil)
	}
	if redirectURI == "" {
		return nil, domain.NewAppError(http.StatusInternalServerError, "AniList login redirect not configured (ANILIST_LOGIN_REDIRECT_URL or APP_URL)", nil)
	}

	tokenResp, err := u.anilist.ExchangeCode(ctx, clientID, clientSecret, redirectURI, strings.TrimSpace(code))
	if err != nil {
		return nil, domain.NewAppError(http.StatusBadRequest, "Failed to exchange Anilist code", err)
	}

	anilistUser, err := u.anilist.GetViewer(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, domain.NewAppError(http.StatusInternalServerError, "Failed to fetch Anilist profile", err)
	}

	user, err := u.userRepo.GetByAnilistID(ctx, anilistUser.ID)
	if err != nil {
		if err == domain.ErrNotFound {
			// Instead of code reuse, we generate a short-lived temp token with the anilist data
			claims := map[string]interface{}{
				"anilist_id":   anilistUser.ID,
				"anilist_name": anilistUser.Name,
				"access_token": tokenResp.AccessToken,
				"refresh_token": tokenResp.RefreshToken,
				"expires_in":   tokenResp.ExpiresIn,
				"type":         "anilist_registration",
			}
			tempToken, _ := u.jwtService.GenerateTempToken(claims, 15*time.Minute)
			
			// Returning the temp token in a custom app error or data
			return nil, domain.NewAppError(428, tempToken, nil)
		}
		return nil, domain.NewAppError(http.StatusInternalServerError, "Database error", err)
	}

	encryptedAccess, err := crypto.Encrypt(tokenResp.AccessToken, u.encryptionKey)
	if err != nil {
		return nil, domain.NewAppError(http.StatusInternalServerError, "Failed to encrypt access token", err)
	}

	encryptedRefresh := ""
	if tokenResp.RefreshToken != "" {
		encryptedRefresh, err = crypto.Encrypt(tokenResp.RefreshToken, u.encryptionKey)
		if err != nil {
			return nil, domain.NewAppError(http.StatusInternalServerError, "Failed to encrypt refresh token", err)
		}
	}
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	identity := &domain.UserSocialIdentity{
		UserID:           user.ID,
		Provider:         "anilist",
		ProviderID:       fmt.Sprintf("%d", anilistUser.ID),
		ProviderUsername: &anilistUser.Name,
		AccessToken:      &encryptedAccess,
		RefreshToken:     &encryptedRefresh,
		ExpiresAt:        &expiresAt,
	}

	if err := u.userRepo.SaveSocialIdentity(ctx, identity); err != nil {
		return nil, domain.NewAppError(http.StatusInternalServerError, "Failed to update user identity", err)
	}

	// Automatic Badge Check
	_ = u.badgeUsecase.CheckAndAwardBadges(ctx, user.ID, "anilist")

	roleSlugs := []string{}
	roles, _ := u.userRepo.GetRolesByUserID(ctx, user.ID)
	for _, r := range roles {
		roleSlugs = append(roleSlugs, r.Slug)
	}
	if len(roleSlugs) == 0 {
		roleSlugs = append(roleSlugs, "user") // Fallback
	}
	user.Roles = roles

	_ = u.xpUsecase.CheckDailyLogin(ctx, user.ID)
	user, _ = u.userRepo.GetByID(ctx, user.ID)

	token, err := u.jwtService.GenerateToken(user.UUID, roleSlugs)
	if err != nil {
		return nil, domain.NewAppError(http.StatusInternalServerError, "Could not generate authentication token", err)
	}

	user.Password = ""
	if user.ScoreFormat == nil || *user.ScoreFormat == "" {
		defaultFormat := "POINT_10_DECIMAL"
		user.ScoreFormat = &defaultFormat
	}
	u.enrichUserImages(user)

	return &AuthTokenResponse{
		Token: token,
		User:  user,
	}, nil
}

func (u *AuthUsecase) autoRegisterGoogleUser(ctx context.Context, googleUser *google.GoogleUser, tokenResp *google.TokenResponse) (*AuthTokenResponse, error) {
	// Generate random password for DB requirements (never used for real login)
	dummyPass, _ := bcrypt.GenerateFromPassword([]byte(uuid.New().String()), bcrypt.DefaultCost)
	sFormat := "POINT_10_DECIMAL"

	newUser := &domain.User{
		UUID:        uuid.New().String(),
		Name:        googleUser.Name,
		Email:       googleUser.Email,
		Password:    string(dummyPass),
		ScoreFormat: &sFormat,
	}

	slug := u.generateUniqueUserSlug(ctx, googleUser.Name)
	newUser.Slug = &slug

	if err := u.userRepo.Create(ctx, newUser); err != nil {
		return nil, domain.NewAppError(http.StatusInternalServerError, "Failed to create account during Google registration", err)
	}

	// 5. Save social identity
	encryptedAccess, _ := crypto.Encrypt(tokenResp.AccessToken, u.encryptionKey)
	encryptedRefresh := ""
	if tokenResp.RefreshToken != "" {
		encryptedRefresh, _ = crypto.Encrypt(tokenResp.RefreshToken, u.encryptionKey)
	}
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	identity := &domain.UserSocialIdentity{
		UserID:           newUser.ID,
		Provider:         "google",
		ProviderID:       googleUser.Sub,
		ProviderUsername: &googleUser.Email,
		AccessToken:      &encryptedAccess,
		RefreshToken:     &encryptedRefresh,
		ExpiresAt:        &expiresAt,
	}

	if err := u.userRepo.SaveSocialIdentity(ctx, identity); err != nil {
		log.Printf("Failed to save google identity after auto-registration: %v", err)
	}

	// Generate avatar in background
	go u.GenerateUserAvatar(context.Background(), newUser.ID, newUser.Name)

	token, _ := u.jwtService.GenerateToken(newUser.UUID, []string{"user"})

	newUser.Password = ""
	u.enrichUserImages(newUser)

	return &AuthTokenResponse{
		Token: token,
		User:  newUser,
	}, nil
}

func (u *AuthUsecase) RegisterWithAnilist(ctx context.Context, tempToken, email string) (*AuthTokenResponse, error) {
	claims, err := u.jwtService.ValidateTempToken(tempToken)
	if err != nil || claims["type"] != "anilist_registration" {
		return nil, domain.NewAppError(http.StatusBadRequest, "Invalid or expired registration session", err)
	}

	anilistID := uint64(claims["anilist_id"].(float64))
	anilistName := claims["anilist_name"].(string)
	accessToken := claims["access_token"].(string)
	refreshToken := claims["refresh_token"].(string)
	expiresIn := int(claims["expires_in"].(float64))

	// Check if email already exists
	existing, _ := u.userRepo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, domain.NewAppError(http.StatusConflict, "A user with this email already exists. Please login and link your account in settings.", nil)
	}

	dummyPass, _ := bcrypt.GenerateFromPassword([]byte(uuid.New().String()), bcrypt.DefaultCost)
	sFormat := "POINT_10_DECIMAL"

	newUser := &domain.User{
		UUID:        uuid.New().String(),
		Name:        anilistName,
		Email:       email,
		Password:    string(dummyPass),
		ScoreFormat: &sFormat,
	}

	slug := u.generateUniqueUserSlug(ctx, anilistName)
	newUser.Slug = &slug

	if err := u.userRepo.Create(ctx, newUser); err != nil {
		return nil, domain.NewAppError(http.StatusInternalServerError, "Failed to create account during AniList registration", err)
	}

	// 5. Save social identity
	encryptedAccess, _ := crypto.Encrypt(accessToken, u.encryptionKey)
	encryptedRefresh := ""
	if refreshToken != "" {
		encryptedRefresh, _ = crypto.Encrypt(refreshToken, u.encryptionKey)
	}
	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)

	identity := &domain.UserSocialIdentity{
		UserID:           newUser.ID,
		Provider:         "anilist",
		ProviderID:       fmt.Sprintf("%d", anilistID),
		ProviderUsername: &anilistName,
		AccessToken:      &encryptedAccess,
		RefreshToken:     &encryptedRefresh,
		ExpiresAt:        &expiresAt,
	}

	if err := u.userRepo.SaveSocialIdentity(ctx, identity); err != nil {
		log.Printf("Failed to save anilist identity after registration: %v", err)
	}

	// Automatic Badge Check
	_ = u.badgeUsecase.CheckAndAwardBadges(ctx, newUser.ID, "anilist")

	go u.GenerateUserAvatar(context.Background(), newUser.ID, newUser.Name)

	token, _ := u.jwtService.GenerateToken(newUser.UUID, []string{"user"})
	newUser.Password = ""
	u.enrichUserImages(newUser)

	return &AuthTokenResponse{
		Token: token,
		User:  newUser,
	}, nil
}

func (u *AuthUsecase) discordRedirectURI() string {
	if ur := strings.TrimSpace(os.Getenv("DISCORD_REDIRECT_URL")); ur != "" {
		return ur
	}
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("APP_URL")), "/")
	if base == "" {
		return ""
	}
	return base + "/settings/account"
}

func (u *AuthUsecase) discordLoginRedirectURI() string {
	if ur := strings.TrimSpace(os.Getenv("DISCORD_LOGIN_REDIRECT_URL")); ur != "" {
		return ur
	}
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("APP_URL")), "/")
	if base == "" {
		return ""
	}
	return base + "/login"
}

func (u *AuthUsecase) DiscordAuthURL(state string) (string, error) {
	clientID := os.Getenv("DISCORD_CLIENT_ID")
	redirectURI := ""
	if state == "discord_login" {
		redirectURI = u.discordLoginRedirectURI()
	} else {
		redirectURI = u.discordRedirectURI()
	}

	if clientID == "" || redirectURI == "" {
		return "", fmt.Errorf("discord OAuth client ID or redirect URI not configured")
	}

	scope := "identify email"
	authURL := fmt.Sprintf(
		"https://discord.com/api/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s",
		url.QueryEscape(clientID),
		url.QueryEscape(redirectURI),
		url.QueryEscape(scope),
		url.QueryEscape(state),
	)
	return authURL, nil
}

func (u *AuthUsecase) LinkDiscord(ctx context.Context, userID uint64, code string) error {
	clientID := os.Getenv("DISCORD_CLIENT_ID")
	clientSecret := os.Getenv("DISCORD_CLIENT_SECRET")
	redirectURI := u.discordRedirectURI()

	tokenResp, err := u.discord.ExchangeCode(ctx, clientID, clientSecret, redirectURI, code)
	if err != nil {
		return domain.NewAppError(http.StatusBadRequest, "Failed to exchange Discord code", err)
	}

	discordUser, err := u.discord.GetUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		return domain.NewAppError(http.StatusInternalServerError, "Failed to fetch Discord profile", err)
	}

	encryptedAccess, _ := crypto.Encrypt(tokenResp.AccessToken, u.encryptionKey)
	var encryptedRefresh *string
	if tokenResp.RefreshToken != "" {
		ref, _ := crypto.Encrypt(tokenResp.RefreshToken, u.encryptionKey)
		encryptedRefresh = &ref
	}
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	identity := &domain.UserSocialIdentity{
		UserID:           userID,
		Provider:         domain.ProviderDiscord,
		ProviderID:       discordUser.ID,
		ProviderUsername: &discordUser.Username,
		AccessToken:      &encryptedAccess,
		RefreshToken:     encryptedRefresh,
		ExpiresAt:        &expiresAt,
	}

	return u.userRepo.SaveSocialIdentity(ctx, identity)
}

func (u *AuthUsecase) LoginWithDiscord(ctx context.Context, code string) (*AuthTokenResponse, error) {
	clientID := os.Getenv("DISCORD_CLIENT_ID")
	clientSecret := os.Getenv("DISCORD_CLIENT_SECRET")
	redirectURI := u.discordLoginRedirectURI()

	tokenResp, err := u.discord.ExchangeCode(ctx, clientID, clientSecret, redirectURI, code)
	if err != nil {
		return nil, domain.NewAppError(http.StatusBadRequest, "Failed to exchange Discord code", err)
	}

	discordUser, err := u.discord.GetUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, domain.NewAppError(http.StatusInternalServerError, "Failed to fetch Discord profile", err)
	}

	// 1. Find identity
	identity, err := u.userRepo.GetSocialIdentity(ctx, domain.ProviderDiscord, discordUser.ID)
	var user *domain.User
	if err != nil {
		if err == domain.ErrNotFound {
			// 2. Try by email if discord is verified
			if discordUser.Verified {
				user, err = u.userRepo.GetByEmail(ctx, discordUser.Email)
				if err != nil && err != domain.ErrNotFound {
					return nil, domain.NewAppError(http.StatusInternalServerError, "Database error", err)
				}
			}

			if user == nil {
				// 3. Auto-register
				user, err = u.autoRegisterDiscordUser(ctx, discordUser)
				if err != nil {
					return nil, err
				}
			}
		} else {
			return nil, domain.NewAppError(http.StatusInternalServerError, "Database error", err)
		}
	} else {
		user, err = u.userRepo.GetByID(ctx, identity.UserID)
		if err != nil {
			return nil, domain.NewAppError(http.StatusInternalServerError, "Database error", err)
		}
	}

	// Update identity with latest tokens
	encryptedAccess, _ := crypto.Encrypt(tokenResp.AccessToken, u.encryptionKey)
	var encryptedRefresh *string
	if tokenResp.RefreshToken != "" {
		ref, _ := crypto.Encrypt(tokenResp.RefreshToken, u.encryptionKey)
		encryptedRefresh = &ref
	}
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	newIdentity := &domain.UserSocialIdentity{
		UserID:           user.ID,
		Provider:         domain.ProviderDiscord,
		ProviderID:       discordUser.ID,
		ProviderUsername: &discordUser.Username,
		AccessToken:      &encryptedAccess,
		RefreshToken:     encryptedRefresh,
		ExpiresAt:        &expiresAt,
	}
	_ = u.userRepo.SaveSocialIdentity(ctx, newIdentity)

	// 5. Generate JWT
	roleSlugs := []string{}
	roles, _ := u.userRepo.GetRolesByUserID(ctx, user.ID)
	for _, r := range roles {
		roleSlugs = append(roleSlugs, r.Slug)
	}
	if len(roleSlugs) == 0 {
		roleSlugs = append(roleSlugs, "user")
	}
	user.Roles = roles

	token, err := u.jwtService.GenerateToken(user.UUID, roleSlugs)
	if err != nil {
		return nil, domain.NewAppError(http.StatusInternalServerError, "Failed to generate token", err)
	}

	user.Password = ""
	u.enrichUserImages(user)

	return &AuthTokenResponse{
		Token: token,
		User:  user,
	}, nil
}

func (u *AuthUsecase) autoRegisterDiscordUser(ctx context.Context, discordUser *discord.DiscordUser) (*domain.User, error) {
	dummyPass, _ := bcrypt.GenerateFromPassword([]byte(uuid.New().String()), bcrypt.DefaultCost)
	sFormat := "POINT_10"

	newUser := &domain.User{
		UUID:        uuid.New().String(),
		Name:        discordUser.Username,
		Email:       discordUser.Email,
		Password:    string(dummyPass),
		ScoreFormat: &sFormat,
	}

	slug := u.generateUniqueUserSlug(ctx, discordUser.Username)
	newUser.Slug = &slug

	if err := u.userRepo.Create(ctx, newUser); err != nil {
		return nil, domain.NewAppError(http.StatusInternalServerError, "Failed to create account during Discord registration", err)
	}

	// Generate avatar in background
	go u.GenerateUserAvatar(context.Background(), newUser.ID, newUser.Name)

	return newUser, nil
}

func (u *AuthUsecase) UnlinkSocial(ctx context.Context, userID uint64, provider string) error {
	// Verify provider is valid
	validProviders := map[string]bool{
		"google":  true,
		"anilist": true,
		"discord": true,
	}

	if !validProviders[provider] {
		return domain.NewAppError(400, "Invalid provider", nil)
	}

	return u.userRepo.DeleteSocialIdentity(ctx, userID, provider)
}
