package v1

import (
	"anirank/api/internal/domain"
	"anirank/api/internal/dto"
	"anirank/api/internal/usecase/auth"
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	usecase *auth.AuthUsecase
}

func NewAuthHandler(u *auth.AuthUsecase) *AuthHandler {
	return &AuthHandler{
		usecase: u,
	}
}

// Login authenticates a user and returns a JWT token.
// @Summary User Login
// @Description Authenticates a user with email and password, returning a JWT.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body object{email=string,password=string} true "User Credentials"
// @Success 200 {object} object{data=string} "Returns JWT token"
// @Failure 400 {object} domain.AppError
// @Failure 401 {object} domain.AppError
// @Failure 422 {object} domain.AppError
// @Router /login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	type LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid JSON body payload", err)
	}

	if req.Email == "" || req.Password == "" {
		return domain.NewAppError(422, "Email and Password are required", nil)
	}

	res, err := h.usecase.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"data": dto.ToAuthResponseDTO(res),
	})
}

// Register registers a new user
// @Summary User Registration
// @Description Registers a new user with name, email, and password.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body object{name=string,email=string,password=string} true "User Data"
// @Success 201 {object} object{data=string} "Returns JWT token"
// @Failure 400 {object} domain.AppError
// @Failure 422 {object} domain.AppError
// @Router /register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	type RegisterRequest struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid JSON body payload", err)
	}

	// Basic Validation
	if len(req.Password) < 8 {
		return domain.NewAppError(422, "Password must be at least 8 characters long", nil)
	}
	if req.Name == "" || req.Email == "" {
		return domain.NewAppError(422, "Required fields are missing", nil)
	}

	res, err := h.usecase.Register(c.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		return err
	}

	return c.Status(201).JSON(fiber.Map{
		"data": dto.ToAuthResponseDTO(res),
	})
}

// Profile checks the JWT Token injected by the middleware
// @Summary Get User Profile
// @Description Retrieves the authenticated user's profile information.
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} object{data=domain.User}
// @Failure 401 {object} domain.AppError
// @Router /profile [get]
func (h *AuthHandler) Profile(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint64)
	if !ok {
		return domain.NewAppError(401, "Corrupted token payload context", nil)
	}

	profile, err := h.usecase.GetProfile(c.Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"data": dto.ToUserDTO(profile),
	})
}

// UpdateAvatar
// @Summary Update User Avatar
// @Description Uploads and updates the user's avatar.
// @Tags Auth
// @Security BearerAuth
// @Accept mpfd
// @Produce json
// @Param image formData file true "Avatar Image"
// @Success 200 {object} object{success=bool,avatar_url=string}
// @Router /users/avatar [post]
func (h *AuthHandler) UpdateAvatar(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint64)
	if !ok {
		return domain.NewAppError(401, "Corrupted token payload context", nil)
	}

	fileHeader, err := c.FormFile("image")
	if err != nil {
		return domain.NewAppError(400, "Image file is required", err)
	}

	if fileHeader.Size > 2*1024*1024 {
		return domain.NewAppError(400, "El avatar no puede superar los 2MB", nil)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return domain.NewAppError(500, "Failed to open image file", err)
	}
	defer file.Close()

	url, err := h.usecase.UpdateAvatar(c.Context(), userID, file, fileHeader.Size, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"avatar_url": url,
		},
	})
}

// UpdateBanner
// @Summary Update User Banner
// @Description Uploads and updates the user's banner.
// @Tags Auth
// @Security BearerAuth
// @Accept mpfd
// @Produce json
// @Param banner formData file true "Banner Image"
// @Success 200 {object} object{success=bool,banner_url=string}
// @Router /users/banner [post]
func (h *AuthHandler) UpdateBanner(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint64)
	if !ok {
		return domain.NewAppError(401, "Corrupted token payload context", nil)
	}

	fileHeader, err := c.FormFile("banner")
	if err != nil {
		return domain.NewAppError(400, "Banner file is required", err)
	}

	if fileHeader.Size > 4*1024*1024 {
		return domain.NewAppError(400, "El banner no puede superar los 4MB", nil)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return domain.NewAppError(500, "Failed to open banner file", err)
	}
	defer file.Close()

	url, err := h.usecase.UpdateBanner(c.Context(), userID, file, fileHeader.Size, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"banner_url": url,
		},
	})
}

// UpdateScoreFormat
// @Summary Update User Score Format
// @Description Updates the user's score format preference.
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body object{score_format=string} true "Score Format"
// @Success 200 {object} object{success=bool}
// @Router /users/score-format [post]
func (h *AuthHandler) UpdateScoreFormat(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint64)
	if !ok {
		return domain.NewAppError(401, "Corrupted token payload context", nil)
	}

	type reqBody struct {
		ScoreFormat any `json:"score_format"`
	}
	var req reqBody
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}

	// Valid score formats whitelist
	validFormats := map[string]bool{
		"POINT_5":          true,
		"POINT_10":         true,
		"POINT_10_DECIMAL": true,
		"POINT_100":        true,
	}

	// Map numeric IDs to format slugs (Matching DB IDs: 1:100, 2:10D, 3:10, 4:5)
	idToFormat := map[int]string{
		1: "POINT_100",
		2: "POINT_10_DECIMAL",
		3: "POINT_10",
		4: "POINT_5",
	}

	var formatStr string
	switch v := req.ScoreFormat.(type) {
	case string:
		formatStr = v
	case float64:
		// JSON numbers are float64 by default
		if mapped, ok := idToFormat[int(v)]; ok {
			formatStr = mapped
		} else {
			formatStr = fmt.Sprintf("%.0f", v)
		}
	case int:
		if mapped, ok := idToFormat[v]; ok {
			formatStr = mapped
		} else {
			formatStr = fmt.Sprintf("%d", v)
		}
	case int64:
		if mapped, ok := idToFormat[int(v)]; ok {
			formatStr = mapped
		} else {
			formatStr = fmt.Sprintf("%d", v)
		}
	}

	if formatStr == "" {
		return domain.NewAppError(422, "score_format is required", nil)
	}

	if !validFormats[formatStr] {
		return domain.NewAppError(422, "Invalid score format. Valid values: POINT_5, POINT_10, POINT_10_DECIMAL, POINT_100", nil)
	}

	if err := h.usecase.UpdateScoreFormat(c.Context(), userID, formatStr); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}

// UpdateProfile
// @Summary Update User Profile Settings
// @Description Updates the user's about text and profile color.
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body object{about=string,profile_color=string} true "Profile Data"
// @Success 200 {object} object{success=bool}
// @Router /users/profile [patch]
func (h *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint64)
	if !ok {
		return domain.NewAppError(401, "Corrupted token payload context", nil)
	}

	type reqBody struct {
		About        *string `json:"about"`
		ProfileColor *string `json:"profile_color"`
	}
	var req reqBody
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}

	if err := h.usecase.UpdateProfile(c.Context(), userID, req.About, req.ProfileColor); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}

func (h *AuthHandler) UpdateEmail(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint64)
	if !ok {
		return domain.NewAppError(401, "Unauthorized", nil)
	}

	var req struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}

	if err := h.usecase.UpdateEmail(c.Context(), userID, req.Email); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Email updated successfully. Please check your new email to verify it.",
	})
}

// AnilistLink redirects to the Anilist authorization page
// @Summary Redirect to Anilist Auth
// @Description Generates the Anilist OAuth2 authorization URL and redirects the user.
// @Tags Auth
// @Security BearerAuth
// @Router /auth/anilist/link [get]
func (h *AuthHandler) AnilistLink(c *fiber.Ctx) error {
	authURL, err := h.usecase.AnilistAuthURL()
	if err != nil {
		return domain.NewAppError(500, "AniList OAuth: "+err.Error(), nil)
	}
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"url": authURL,
		},
	})
}

// AnilistCallback handles the code from Anilist and links the account
// @Summary Anilist OAuth2 Callback
// @Description Receives the authorization code from Anilist and links it to the authenticated user.
// @Tags Auth
// @Security BearerAuth
// @Param request body object{code=string} true "Auth Code"
// @Router /auth/anilist/callback [post]
func (h *AuthHandler) AnilistCallback(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint64)
	if !ok {
		return domain.NewAppError(401, "Unauthorized", nil)
	}

	type reqBody struct {
		Code string `json:"code"`
	}
	var req reqBody
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}

	if req.Code == "" {
		return domain.NewAppError(422, "code is required", nil)
	}

	if err := h.usecase.LinkAnilist(c.Context(), userID, req.Code); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "Anilist account linked successfully",
		},
	})
}

func (h *AuthHandler) GoogleLink(c *fiber.Ctx) error {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	redirectURI := os.Getenv("GOOGLE_REDIRECT_URL")
	scope := "openid email profile"

	url := fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&access_type=offline&prompt=consent",
		clientID, redirectURI, scope)

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"url": url,
		},
	})
}

func (h *AuthHandler) GoogleCallback(c *fiber.Ctx) error {
	var req struct {
		Code string `json:"code"`
	}
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(fiber.StatusBadRequest, "Invalid request body", err)
	}

	userID := c.Locals("user_id").(uint64)
	if err := h.usecase.LinkGoogle(c.Context(), userID, req.Code); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "Google account linked successfully",
		},
	})
}

func (h *AuthHandler) GoogleLogin(c *fiber.Ctx) error {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	redirectURI := h.getGoogleLoginRedirectURI()

	scope := "openid email profile"
	url := fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&access_type=offline&prompt=consent",
		clientID, redirectURI, scope)

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"url": url,
		},
	})
}

func (h *AuthHandler) GoogleLoginCallback(c *fiber.Ctx) error {
	var req struct {
		Code string `json:"code"`
	}
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(fiber.StatusBadRequest, "Invalid request body", err)
	}

	redirectURI := h.getGoogleLoginRedirectURI()
	res, err := h.usecase.LoginWithGoogle(c.Context(), req.Code, redirectURI)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"data": dto.ToAuthResponseDTO(res),
	})
}

func (h *AuthHandler) AnilistLogin(c *fiber.Ctx) error {
	authURL, err := h.usecase.AnilistLoginAuthURL()
	if err != nil {
		return domain.NewAppError(500, "AniList OAuth: "+err.Error(), nil)
	}
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"url": authURL,
		},
	})
}

func (h *AuthHandler) AnilistLoginCallback(c *fiber.Ctx) error {
	var req struct {
		Code string `json:"code"`
	}
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(fiber.StatusBadRequest, "Invalid request body", err)
	}

	res, err := h.usecase.LoginWithAnilist(c.Context(), req.Code)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"data": dto.ToAuthResponseDTO(res),
	})
}

func (h *AuthHandler) AnilistRegister(c *fiber.Ctx) error {
	var req struct {
		TempToken string `json:"tempToken"`
		Email     string `json:"email"`
	}
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(fiber.StatusBadRequest, "Invalid request body", err)
	}

	res, err := h.usecase.RegisterWithAnilist(c.Context(), req.TempToken, req.Email)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"data": dto.ToAuthResponseDTO(res),
	})
}

func (h *AuthHandler) DiscordLink(c *fiber.Ctx) error {
	authURL, err := h.usecase.DiscordAuthURL("discord_link")
	if err != nil {
		return domain.NewAppError(500, "Discord OAuth: "+err.Error(), nil)
	}
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"url": authURL,
		},
	})
}

func (h *AuthHandler) DiscordCallback(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint64)
	if !ok {
		return domain.NewAppError(401, "Unauthorized", nil)
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid JSON body", err)
	}

	if req.Code == "" {
		return domain.NewAppError(422, "code is required", nil)
	}

	if err := h.usecase.LinkDiscord(c.Context(), userID, req.Code); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "Discord account linked successfully",
		},
	})
}

func (h *AuthHandler) DiscordLogin(c *fiber.Ctx) error {
	authURL, err := h.usecase.DiscordAuthURL("discord_login")
	if err != nil {
		return domain.NewAppError(500, "Discord OAuth: "+err.Error(), nil)
	}
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"url": authURL,
		},
	})
}

func (h *AuthHandler) DiscordLoginCallback(c *fiber.Ctx) error {
	var req struct {
		Code string `json:"code"`
	}
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid JSON body", err)
	}

	if req.Code == "" {
		return domain.NewAppError(422, "code is required", nil)
	}

	res, err := h.usecase.LoginWithDiscord(c.Context(), req.Code)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"data": dto.ToAuthResponseDTO(res),
	})
}

func (h *AuthHandler) getGoogleLoginRedirectURI() string {
	redirectURI := os.Getenv("GOOGLE_LOGIN_REDIRECT_URL")
	if redirectURI != "" {
		return redirectURI
	}

	base := os.Getenv("GOOGLE_REDIRECT_URL")
	if strings.Contains(base, "/settings/account") {
		return strings.Replace(base, "/settings/account", "/login", 1)
	}
	return base
}

// UnlinkSocial removes a social identity from the user
// @Summary Unlink Social Identity
// @Description Removes the specified social provider linkage from the authenticated user.
// @Tags Auth
// @Security BearerAuth
// @Param provider path string true "Provider (google, anilist, discord)"
// @Success 200 {object} object{success=bool}
// @Router /auth/{provider}/unlink [delete]
func (h *AuthHandler) UnlinkSocial(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint64)
	if !ok {
		return domain.NewAppError(401, "Unauthorized", nil)
	}

	provider := c.Params("provider")
	if provider == "" {
		return domain.NewAppError(422, "provider is required", nil)
	}

	if err := h.usecase.UnlinkSocial(c.Context(), userID, provider); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("%s account unlinked successfully", strings.Title(provider)),
	})
}

func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return domain.NewAppError(400, "Verification token is required", nil)
	}

	if err := h.usecase.VerifyEmail(c.Context(), token); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"message": "Email verified successfully",
	})
}

func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	type ForgotRequest struct {
		Email string `json:"email"`
	}

	var req ForgotRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid JSON body payload", err)
	}

	if req.Email == "" {
		return domain.NewAppError(422, "Email is required", nil)
	}

	if err := h.usecase.ForgotPassword(c.Context(), req.Email); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"message": "If an account with that email exists, we have sent instructions to reset your password",
	})
}

func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	type ResetRequest struct {
		Token       string `json:"token"`
		NewPassword string `json:"password"`
	}

	var req ResetRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid JSON body payload", err)
	}

	if req.Token == "" || req.NewPassword == "" {
		return domain.NewAppError(422, "Token and password are required", nil)
	}

	if len(req.NewPassword) < 8 {
		return domain.NewAppError(422, "Password must be at least 8 characters long", nil)
	}

	if err := h.usecase.ResetPassword(c.Context(), req.Token, req.NewPassword); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"message": "Password has been successfully reset",
	})
}
