package testutil

import (
	"context"
	"io"
	"os"
	"time"

	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"
	"anirank/api/internal/infrastructure/anilist"
	"anirank/api/internal/infrastructure/discord"
	"anirank/api/internal/infrastructure/google"
	"anirank/api/internal/usecase/auth"

	"github.com/gofiber/fiber/v2"
)

// MockUserRepository implements domain.UserRepository for testing
type MockUserRepository struct {
	User           *domain.User
	Err            error
	GetByIDFunc    func(id uint64) (*domain.User, error)
	GetByUUIDFunc  func(uuid string) (*domain.User, error)
	GetRolesFunc   func() ([]domain.Role, error)
	GetPermsFunc   func(roleID uint64) ([]domain.Permission, error)
	GetByEmailFunc func(email string) (*domain.User, error)
	GetByGoogleIDFunc func(googleID string) (*domain.User, error)
	SaveSocialIdentityFunc func(identity *domain.UserSocialIdentity) error
	CreateFunc             func(user *domain.User) error
	GetLastInteractionTimeFunc func(userID uint64) (time.Time, error)
	UpdateSoftbanStatusFunc    func(userID uint64, status bool) error
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uint64) (*domain.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return m.User, m.Err
}

func (m *MockUserRepository) GetByUUID(ctx context.Context, uuid string) (*domain.User, error) {
	if m.GetByUUIDFunc != nil {
		return m.GetByUUIDFunc(uuid)
	}
	return m.User, m.Err
}

func (m *MockUserRepository) GetRoles(ctx context.Context) ([]domain.Role, error) {
	if m.GetRolesFunc != nil {
		return m.GetRolesFunc()
	}
	return nil, nil
}

func (m *MockUserRepository) GetPermissionsByRoleID(ctx context.Context, roleID uint64) ([]domain.Permission, error) {
	if m.GetPermsFunc != nil {
		return m.GetPermsFunc(roleID)
	}
	return nil, nil
}

// Implement other methods as no-ops for now to satisfy interface
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(email)
	}
	return m.User, m.Err
}
func (m *MockUserRepository) GetByGoogleID(ctx context.Context, googleID string) (*domain.User, error) {
	if m.GetByGoogleIDFunc != nil {
		return m.GetByGoogleIDFunc(googleID)
	}
	return m.User, m.Err
}
func (m *MockUserRepository) GetByAnilistID(ctx context.Context, alID uint64) (*domain.User, error)    { return m.User, m.Err }
func (m *MockUserRepository) GetBySlug(ctx context.Context, slug string) (*domain.User, error)         { return m.User, m.Err }
func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(user)
	}
	return m.Err
}
func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error                      { return m.Err }
func (m *MockUserRepository) Delete(ctx context.Context, id uint64) error                             { return m.Err }
func (m *MockUserRepository) GetRolesByUserID(ctx context.Context, userID uint64) ([]domain.Role, error) { return nil, nil }
func (m *MockUserRepository) UpdateRoles(ctx context.Context, userID uint64, rIDs []uint64) error      { return nil }
func (m *MockUserRepository) GetBadgesByUserID(ctx context.Context, userID uint64) ([]domain.Badge, error) { return nil, nil }
func (m *MockUserRepository) GetBadgesByUserIDs(ctx context.Context, userIDs []uint64) (map[uint64][]domain.Badge, error) { return nil, nil }
func (m *MockUserRepository) UpdateBadges(ctx context.Context, userID uint64, bIDs []uint64) error     { return nil }
func (m *MockUserRepository) GetPermissionsByUserID(ctx context.Context, userID uint64) ([]domain.Permission, error) { return nil, nil }
func (m *MockUserRepository) GetAllPermissions(ctx context.Context) ([]domain.Permission, error)      { return nil, nil }
func (m *MockUserRepository) UpdateRolePermissions(ctx context.Context, rID uint64, pIDs []uint64) error { return nil }
func (m *MockUserRepository) SetImage(ctx context.Context, uID uint64, iT, iP string) error            { return nil }
func (m *MockUserRepository) UpdateScoreFormat(ctx context.Context, uID uint64, f string) error      { return nil }
func (m *MockUserRepository) UpdatePassword(ctx context.Context, uID uint64, hP string) error         { return nil }
func (m *MockUserRepository) GetUsers(ctx context.Context, p, l int, s string) ([]domain.User, int, error) { return nil, 0, nil }
func (m *MockUserRepository) GetRanking(ctx context.Context, sB string, l, o int) ([]domain.RankingUser, int, error) { return nil, 0, nil }
func (m *MockUserRepository) Follow(ctx context.Context, fI, fdI uint64) error                         { return nil }
func (m *MockUserRepository) Unfollow(ctx context.Context, fI, fdI uint64) error                       { return nil }
func (m *MockUserRepository) IsFollowing(ctx context.Context, fI, fdI uint64) (bool, error)            { return false, nil }
func (m *MockUserRepository) GetFollowersCount(ctx context.Context, uID uint64) (int, error)            { return 0, nil }
func (m *MockUserRepository) GetFollowingCount(ctx context.Context, uID uint64) (int, error)            { return 0, nil }
func (m *MockUserRepository) GetFollowers(ctx context.Context, uID uint64, l, o int) ([]domain.User, error) { return nil, nil }
func (m *MockUserRepository) GetFollowing(ctx context.Context, uID uint64, l, o int) ([]domain.User, error) { return nil, nil }
func (m *MockUserRepository) GetMany(ctx context.Context, ids []uint64) ([]domain.User, error)         { return nil, nil }
func (m *MockUserRepository) GetLastInteractionTime(ctx context.Context, userID uint64) (time.Time, error) {
	if m.GetLastInteractionTimeFunc != nil {
		return m.GetLastInteractionTimeFunc(userID)
	}
	return time.Time{}, nil
}
func (m *MockUserRepository) UpdateSoftbanStatus(ctx context.Context, userID uint64, status bool) error {
	if m.UpdateSoftbanStatusFunc != nil {
		return m.UpdateSoftbanStatusFunc(userID, status)
	}
	return nil
}
func (m *MockUserRepository) GetSocialIdentity(ctx context.Context, provider, providerID string) (*domain.UserSocialIdentity, error) {
	return nil, nil
}
func (m *MockUserRepository) GetSocialIdentitiesByUserID(ctx context.Context, userID uint64) ([]domain.UserSocialIdentity, error) {
	return nil, nil
}
func (m *MockUserRepository) SaveSocialIdentity(ctx context.Context, identity *domain.UserSocialIdentity) error {
	if m.SaveSocialIdentityFunc != nil {
		return m.SaveSocialIdentityFunc(identity)
	}
	return nil
}
func (m *MockUserRepository) DeleteSocialIdentity(ctx context.Context, userID uint64, provider string) error {
	return nil
}

// MockStorageService implements infrastructure.StorageService
type MockStorageService struct{}
func (m *MockStorageService) UploadFile(ctx context.Context, rP string, f io.Reader, s int64, cT string) (string, error) { return "", nil }
func (m *MockStorageService) GetURL(rP string) string { return rP }
func (m *MockStorageService) DeleteFile(ctx context.Context, rP string) error { return nil }
func (m *MockStorageService) GetFile(ctx context.Context, rP string) (io.ReadCloser, error) { return nil, nil }
func (m *MockStorageService) FileExists(ctx context.Context, rP string) (bool, error) { return false, nil }
func (m *MockStorageService) ListFiles(ctx context.Context, p string) ([]string, error) { return nil, nil }
func (m *MockStorageService) GetEndpoint() string { return "http://mock-storage" }
func (m *MockStorageService) GetPublicURL() string { return "http://public-url" }

// MockCache implements domain.Cache
type MockCache struct {
	Data map[string]interface{}
}

func (m *MockCache) Get(ctx context.Context, key string, dest interface{}) error { return nil }
func (m *MockCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error { return nil }
func (m *MockCache) Delete(ctx context.Context, key string) error { return nil }
func (m *MockCache) IsAvailable() bool { return true }
func (m *MockCache) Publish(ctx context.Context, channel string, message interface{}) error { return nil }
func (m *MockCache) Subscribe(ctx context.Context, channel string) (domain.Subscriber, error) { return nil, nil }

// MockMediaService implements infrastructure.MediaService
type MockMediaService struct{}
func (m *MockMediaService) GetURL(p string) string { return p }
func (m *MockMediaService) Resolve(p *string) *string { return p }
func (m *MockMediaService) GeneratePath(pr string, id uint64, e string) string { return "" }
func (m *MockMediaService) UploadImage(ctx context.Context, p string, id uint64, f io.Reader, s int64, cT string) (string, string, error) { return "", "", nil }
func (m *MockMediaService) UploadImageOptimized(ctx context.Context, p string, id uint64, f io.Reader, o infrastructure.ImageOptions) (string, string, error) { return "", "", nil }
func (m *MockMediaService) UploadVideo(ctx context.Context, p string, id uint64, f io.Reader, s int64, cT string, oN string) (string, string, error) { return "", "", nil }
func (m *MockMediaService) UploadWithResolutions(ctx context.Context, p string, id uint64, f io.Reader, pr infrastructure.ResolutionPreset) (string, string, error) { return "", "", nil }
func (m *MockMediaService) GetImageSources(p string) []domain.ImageSource { return nil }
func (m *MockMediaService) GetFile(ctx context.Context, p string) (io.ReadCloser, error) { return nil, nil }
func (m *MockMediaService) DeleteMedia(ctx context.Context, p string) {}
func (m *MockMediaService) FileExists(ctx context.Context, p string) (bool, error) { return false, nil }

// Gamification Mocks
type MockXPUsecase struct{}
func (m *MockXPUsecase) AwardXP(ctx context.Context, uID uint64, aK string, meta map[string]interface{}) error { return nil }
func (m *MockXPUsecase) CheckDailyLogin(ctx context.Context, uID uint64) error { return nil }

type MockBadgeUsecase struct{}
func (m *MockBadgeUsecase) GetByID(ctx context.Context, id uint64) (*domain.Badge, error) { return nil, nil }
func (m *MockBadgeUsecase) GetAll(ctx context.Context) ([]domain.Badge, error) { return nil, nil }
func (m *MockBadgeUsecase) Create(ctx context.Context, b *domain.Badge, meta domain.AuditMetadata) error { return nil }
func (m *MockBadgeUsecase) Update(ctx context.Context, b *domain.Badge, meta domain.AuditMetadata) error { return nil }
func (m *MockBadgeUsecase) Delete(ctx context.Context, id uint64, meta domain.AuditMetadata) error { return nil }
func (m *MockBadgeUsecase) HandleBadgeIcon(c *fiber.Ctx, b *domain.Badge) error { return nil }
func (m *MockBadgeUsecase) ResolveBadgeURL(b *domain.Badge) {}
func (m *MockBadgeUsecase) ResolveBadgesURLs(b []domain.Badge) {}
func (m *MockBadgeUsecase) CheckAndAwardBadges(ctx context.Context, uID uint64, type_ string) error { return nil }
func (m *MockBadgeUsecase) ProcessBadgeIcons(ctx context.Context, progress chan<- string) error { return nil }

// CreateTestToken generates a valid JWT for testing
func CreateTestToken(userUUID string, roles []string) (string, error) {
	// Use a fixed secret for testing
	os.Setenv("JWT_SECRET", "test_secret_key_12345")
	jwtSvc := auth.NewJWTService()
	return jwtSvc.GenerateToken(userUUID, roles)
}

// CreateExpiredTestToken generates a JWT that is already expired
func CreateExpiredTestToken(userUUID string, roles []string) (string, error) {
	os.Setenv("JWT_SECRET", "test_secret_key_12345")
	// This is a bit hacky because JWTService doesn't expose a way to set custom expiration easily
	// but we can just use the internal GenerateTempToken with negative duration
	jwtSvc := auth.NewJWTService()
	data := map[string]interface{}{
		"user_uuid": userUUID,
		"roles":     roles,
	}
	return jwtSvc.GenerateTempToken(data, -1*time.Hour)
}

// NewTestApp returns a fiber app configured for testing with an AppError handler
func NewTestApp() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if e, ok := err.(*domain.AppError); ok {
				return c.Status(e.Code).JSON(fiber.Map{"message": e.Message})
			}
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		},
	})
}
// OAuth Mocks
type MockAnilistClient struct {
	ExchangeCodeFunc    func(clientID, clientSecret, redirectURI, code string) (*anilist.TokenResponse, error)
	GetViewerFunc       func(accessToken string) (*anilist.AnilistUser, error)
}
func (m *MockAnilistClient) SearchAnimes(ctx context.Context, s string, f string, p int) (*anilist.AnilistResponse, error) { return nil, nil }
func (m *MockAnilistClient) GetMediaByIDs(ctx context.Context, ids []int) ([]anilist.Media, error) { return nil, nil }
func (m *MockAnilistClient) FetchAnimes(ctx context.Context, p int, s string, sy int, f string) (*anilist.AnilistResponse, error) { return nil, nil }
func (m *MockAnilistClient) GetUserMediaList(ctx context.Context, aID int64, s string, p, pp int) (*anilist.AnilistMediaListResponse, error) { return nil, nil }
func (m *MockAnilistClient) ExchangeCode(ctx context.Context, clientID, clientSecret, redirectURI, code string) (*anilist.TokenResponse, error) {
	if m.ExchangeCodeFunc != nil {
		return m.ExchangeCodeFunc(clientID, clientSecret, redirectURI, code)
	}
	return &anilist.TokenResponse{AccessToken: "test-token"}, nil
}
func (m *MockAnilistClient) GetViewer(ctx context.Context, accessToken string) (*anilist.AnilistUser, error) {
	if m.GetViewerFunc != nil {
		return m.GetViewerFunc(accessToken)
	}
	return &anilist.AnilistUser{ID: 123, Name: "AnilistUser"}, nil
}
func (m *MockAnilistClient) SearchStaff(ctx context.Context, s string) ([]anilist.Staff, error) { return nil, nil }
func (m *MockAnilistClient) SearchStaffBatch(ctx context.Context, r []anilist.StaffSearchReq) (map[string][]anilist.Staff, error) { return nil, nil }
func (m *MockAnilistClient) Ping(ctx context.Context) error { return nil }

type MockGoogleClient struct {
	ExchangeCodeFunc func(clientID, clientSecret, redirectURI, code string) (*google.TokenResponse, error)
	GetUserInfoFunc  func(accessToken string) (*google.GoogleUser, error)
}
func (m *MockGoogleClient) ExchangeCode(ctx context.Context, clientID, clientSecret, redirectURI, code string) (*google.TokenResponse, error) {
	if m.ExchangeCodeFunc != nil {
		return m.ExchangeCodeFunc(clientID, clientSecret, redirectURI, code)
	}
	return &google.TokenResponse{AccessToken: "google-token", ExpiresIn: 3600}, nil
}
func (m *MockGoogleClient) GetUserInfo(ctx context.Context, accessToken string) (*google.GoogleUser, error) {
	if m.GetUserInfoFunc != nil {
		return m.GetUserInfoFunc(accessToken)
	}
	return &google.GoogleUser{Sub: "g-123", Email: "google@test.com", Name: "Google User", EmailVerified: true}, nil
}

type MockDiscordClient struct {
	ExchangeCodeFunc func(clientID, clientSecret, redirectURI, code string) (*discord.TokenResponse, error)
	GetUserInfoFunc  func(accessToken string) (*discord.DiscordUser, error)
}
func (m *MockDiscordClient) ExchangeCode(ctx context.Context, clientID, clientSecret, redirectURI, code string) (*discord.TokenResponse, error) {
	if m.ExchangeCodeFunc != nil {
		return m.ExchangeCodeFunc(clientID, clientSecret, redirectURI, code)
	}
	return &discord.TokenResponse{AccessToken: "discord-token", ExpiresIn: 3600}, nil
}
func (m *MockDiscordClient) GetUserInfo(ctx context.Context, accessToken string) (*discord.DiscordUser, error) {
	if m.GetUserInfoFunc != nil {
		return m.GetUserInfoFunc(accessToken)
	}
	return &discord.DiscordUser{ID: "d-123", Username: "DiscordUser", Email: "discord@test.com"}, nil
}

type MockMailService struct {
	SendVerificationFunc  func(to, name, token string) error
	SendPasswordResetFunc func(to, name, token string) error
}

func (m *MockMailService) SendVerificationEmail(ctx context.Context, to string, name string, token string) error {
	if m.SendVerificationFunc != nil {
		return m.SendVerificationFunc(to, name, token)
	}
	return nil
}

func (m *MockMailService) SendPasswordResetEmail(ctx context.Context, to string, name string, token string) error {
	if m.SendPasswordResetFunc != nil {
		return m.SendPasswordResetFunc(to, name, token)
	}
	return nil
}

type MockAuthTokenRepository struct {
	CreateFunc     func(token *domain.AuthToken) error
	GetByTokenFunc func(token string, tokenType string) (*domain.AuthToken, error)
}

func (m *MockAuthTokenRepository) Create(ctx context.Context, token *domain.AuthToken) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(token)
	}
	return nil
}

func (m *MockAuthTokenRepository) GetByToken(ctx context.Context, token string, tokenType string) (*domain.AuthToken, error) {
	if m.GetByTokenFunc != nil {
		return m.GetByTokenFunc(token, tokenType)
	}
	return &domain.AuthToken{}, nil
}

func (m *MockAuthTokenRepository) DeleteByUser(ctx context.Context, userID uint64, tokenType string) error {
	return nil
}

func (m *MockAuthTokenRepository) Delete(ctx context.Context, id uint64) error {
	return nil
}

func (m *MockAuthTokenRepository) CleanupExpired(ctx context.Context) error {
	return nil
}
