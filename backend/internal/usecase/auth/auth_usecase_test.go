package auth_test

import (
	"context"
	"os"
	"testing"

	"anirank/api/internal/domain"
	"anirank/api/internal/testutil"
	"anirank/api/internal/usecase/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthUsecase_Login(t *testing.T) {
	ctx := context.Background()
	password := "securepassword123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	
	testUser := &domain.User{
		ID:       1,
		UUID:     "user-1",
		Email:    "test@example.com",
		Password: string(hash),
	}

	mockRepo := &testutil.MockUserRepository{User: testUser}
	jwtSvc := auth.NewJWTService()
	xpMock := &testutil.MockXPUsecase{}
	mailMock := &testutil.MockMailService{}
	tokenRepoMock := &testutil.MockAuthTokenRepository{}
	
	uc := auth.NewAuthUsecase(mockRepo, jwtSvc, &testutil.MockStorageService{}, &testutil.MockMediaService{}, xpMock, &testutil.MockBadgeUsecase{}, &testutil.MockAnilistClient{}, &testutil.MockGoogleClient{}, &testutil.MockDiscordClient{}, mailMock, tokenRepoMock, "test-key-32-chars-long-1234567890")

	t.Run("Success", func(t *testing.T) {
		resp, err := uc.Login(ctx, testUser.Email, password)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, testUser.UUID, resp.User.UUID)
		assert.Empty(t, resp.User.Password) // Password should be scrubbed
	})

	t.Run("Invalid Password", func(t *testing.T) {
		resp, err := uc.Login(ctx, testUser.Email, "wrongpassword")
		assert.Error(t, err)
		assert.Nil(t, resp)
		appErr := err.(*domain.AppError)
		assert.Equal(t, 401, appErr.Code)
	})

	t.Run("User Not Found", func(t *testing.T) {
		mockRepo.User = nil
		mockRepo.Err = domain.ErrNotFound
		resp, err := uc.Login(ctx, "nonexistent@test.com", password)
		assert.Error(t, err)
		assert.Nil(t, resp)
		appErr := err.(*domain.AppError)
		assert.Equal(t, 401, appErr.Code)
	})
}

func TestAuthUsecase_Register(t *testing.T) {
	ctx := context.Background()
	mockRepo := &testutil.MockUserRepository{User: nil} // Simulate user doesn't exist initially
	jwtSvc := auth.NewJWTService()
	mailMock := &testutil.MockMailService{}
	tokenRepoMock := &testutil.MockAuthTokenRepository{}
	
	uc := auth.NewAuthUsecase(mockRepo, jwtSvc, &testutil.MockStorageService{}, &testutil.MockMediaService{}, &testutil.MockXPUsecase{}, &testutil.MockBadgeUsecase{}, &testutil.MockAnilistClient{}, &testutil.MockGoogleClient{}, &testutil.MockDiscordClient{}, mailMock, tokenRepoMock, "test-key-32-chars-long-1234567890")

	t.Run("Success", func(t *testing.T) {
		// Mock unique slug generation logic if needed, but generateUniqueUserSlug is internal
		resp, err := uc.Register(ctx, "New User", "new@test.com", "pass123")
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "new@test.com", resp.User.Email)
		assert.NotEmpty(t, resp.Token)
	})

	t.Run("Duplicate Email", func(t *testing.T) {
		mockRepo.User = &domain.User{ID: 1, Email: "existing@test.com"}
		resp, err := uc.Register(ctx, "Other User", "existing@test.com", "pass123")
		assert.Error(t, err)
		assert.Nil(t, resp)
		appErr := err.(*domain.AppError)
		assert.Equal(t, 400, appErr.Code)
		assert.Contains(t, appErr.Message, "already registered")
	})
}

func TestAuthUsecase_LinkAnilist(t *testing.T) {
	ctx := context.Background()
	mockRepo := &testutil.MockUserRepository{}
	
	// Track social identities saved
	var savedIdentity *domain.UserSocialIdentity
	mockRepo.SaveSocialIdentityFunc = func(identity *domain.UserSocialIdentity) error {
		savedIdentity = identity
		return nil
	}

	os.Setenv("ANILIST_CLIENT_ID", "test-id")
	os.Setenv("ANILIST_CLIENT_SECRET", "test-secret")
	os.Setenv("APP_URL", "http://localhost")

	encryptionKey := "abc123abc123abc123abc123abc123ab" // 32 chars
	mailMock := &testutil.MockMailService{}
	tokenRepoMock := &testutil.MockAuthTokenRepository{}
	uc := auth.NewAuthUsecase(mockRepo, auth.NewJWTService(), &testutil.MockStorageService{}, &testutil.MockMediaService{}, &testutil.MockXPUsecase{}, &testutil.MockBadgeUsecase{}, &testutil.MockAnilistClient{}, &testutil.MockGoogleClient{}, &testutil.MockDiscordClient{}, mailMock, tokenRepoMock, encryptionKey)

	t.Run("Success - Verify Encryption", func(t *testing.T) {
		err := uc.LinkAnilist(ctx, 1, "test-code")
		assert.NoError(t, err)
		require.NotNil(t, savedIdentity)
		assert.Equal(t, uint64(1), savedIdentity.UserID)
		assert.Equal(t, "anilist", savedIdentity.Provider)
		assert.NotEmpty(t, *savedIdentity.AccessToken)
		
		// Token should be encrypted (should not be plain "test-token" from the mock)
		assert.NotEqual(t, "test-token", *savedIdentity.AccessToken)
	})
}

func TestAuthUsecase_LoginWithGoogle(t *testing.T) {
	os.Setenv("GOOGLE_CLIENT_ID", "test-google-id")
	os.Setenv("GOOGLE_CLIENT_SECRET", "test-google-secret")
	os.Setenv("GOOGLE_REDIRECT_URL", "http://localhost/callback")

	ctx := context.Background()
	mockRepo := &testutil.MockUserRepository{}
	
	mailMock := &testutil.MockMailService{}
	tokenRepoMock := &testutil.MockAuthTokenRepository{}
	uc := auth.NewAuthUsecase(mockRepo, auth.NewJWTService(), &testutil.MockStorageService{}, &testutil.MockMediaService{}, &testutil.MockXPUsecase{}, &testutil.MockBadgeUsecase{}, &testutil.MockAnilistClient{}, &testutil.MockGoogleClient{}, &testutil.MockDiscordClient{}, mailMock, tokenRepoMock, "test-key-32-chars-long-1234567890")

	t.Run("Login Existing User", func(t *testing.T) {
		mockRepo.User = &domain.User{ID: 10, UUID: "google-user-uuid"}
		mockRepo.Err = nil
		
		resp, err := uc.LoginWithGoogle(ctx, "google-code", "http://localhost/callback")
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "google-user-uuid", resp.User.UUID)
	})

	t.Run("Auto-Register New User", func(t *testing.T) {
		mockRepo.User = nil
		mockRepo.Err = nil // Ensure Create succeeds
		
		// Map GetByGoogleID to Not Found
		mockRepo.GetByGoogleIDFunc = func(gID string) (*domain.User, error) {
			return nil, domain.ErrNotFound
		}
		// Map GetByEmail to Not Found
		mockRepo.GetByEmailFunc = func(email string) (*domain.User, error) {
			return nil, domain.ErrNotFound
		}
		mockRepo.CreateFunc = func(user *domain.User) error {
			user.ID = 1
			return nil
		}

		resp, err := uc.LoginWithGoogle(ctx, "google-code", "http://localhost/callback")
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "Google User", resp.User.Name)
		assert.Equal(t, "google@test.com", resp.User.Email)
	})
}

func TestAuthUsecase_LinkSocials(t *testing.T) {
	ctx := context.Background()
	mockRepo := &testutil.MockUserRepository{}
	encryptionKey := "abc123abc123abc123abc123abc123ab"
	mailMock := &testutil.MockMailService{}
	tokenRepoMock := &testutil.MockAuthTokenRepository{}
	uc := auth.NewAuthUsecase(mockRepo, auth.NewJWTService(), &testutil.MockStorageService{}, &testutil.MockMediaService{}, &testutil.MockXPUsecase{}, &testutil.MockBadgeUsecase{}, &testutil.MockAnilistClient{}, &testutil.MockGoogleClient{}, &testutil.MockDiscordClient{}, mailMock, tokenRepoMock, encryptionKey)

	os.Setenv("GOOGLE_CLIENT_ID", "test")
	os.Setenv("GOOGLE_CLIENT_SECRET", "test")
	os.Setenv("GOOGLE_REDIRECT_URL", "http://localhost/callback")
	os.Setenv("DISCORD_CLIENT_ID", "test")
	os.Setenv("DISCORD_CLIENT_SECRET", "test")
	os.Setenv("APP_URL", "http://localhost")

	t.Run("Link Google", func(t *testing.T) {
		var savedIdentity *domain.UserSocialIdentity
		mockRepo.SaveSocialIdentityFunc = func(identity *domain.UserSocialIdentity) error {
			savedIdentity = identity
			return nil
		}

		err := uc.LinkGoogle(ctx, 1, "test-code")
		assert.NoError(t, err)
		require.NotNil(t, savedIdentity)
		assert.Equal(t, "google", savedIdentity.Provider)
		assert.NotEqual(t, "test-token", *savedIdentity.AccessToken) // Encrypted
	})

	t.Run("Link Discord", func(t *testing.T) {
		var savedIdentity *domain.UserSocialIdentity
		mockRepo.SaveSocialIdentityFunc = func(identity *domain.UserSocialIdentity) error {
			savedIdentity = identity
			return nil
		}

		err := uc.LinkDiscord(ctx, 1, "test-code")
		assert.NoError(t, err)
		require.NotNil(t, savedIdentity)
		assert.Equal(t, "discord", savedIdentity.Provider)
	})
}
