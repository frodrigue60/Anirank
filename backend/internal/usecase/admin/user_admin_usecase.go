package admin

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"strings"

	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"
	"anirank/api/internal/pkg/utils"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserAdminUsecase struct {
	userRepo     domain.UserRepository
	mediaService infrastructure.MediaService
	auditUsecase domain.AuditLogUsecase
}

func NewUserAdminUsecase(
	ur domain.UserRepository,
	media infrastructure.MediaService,
	audit domain.AuditLogUsecase,
) *UserAdminUsecase {
	return &UserAdminUsecase{
		userRepo:     ur,
		mediaService: media,
		auditUsecase: audit,
	}
}

// ---- USERS ----
func (u *UserAdminUsecase) GetUsers(ctx context.Context, page, limit int, search string) ([]domain.User, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	users, total, err := u.userRepo.GetUsers(ctx, page, limit, search)
	if err == nil {
		u.ResolveUsersURLs(users)
	}
	return users, total, err
}

func (u *UserAdminUsecase) GetRoles(ctx context.Context) ([]domain.Role, error) {
	return u.userRepo.GetRoles(ctx)
}

func (u *UserAdminUsecase) GetUser(ctx context.Context, id uint64) (*domain.User, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	u.ResolveUserURLs(user)

	// Load roles and badges
	roles, _ := u.userRepo.GetRolesByUserID(ctx, id)
	user.Roles = roles

	badges, _ := u.userRepo.GetBadgesByUserID(ctx, id)
	// Resolve badge icons if necessary
	if u.mediaService != nil {
		for i := range badges {
			if badges[i].Icon != nil && *badges[i].Icon != "" {
				url := u.mediaService.Resolve(badges[i].Icon)
				badges[i].IconUrl = url
			}
		}
	}
	user.Badges = badges

	return user, nil
}

func (u *UserAdminUsecase) CreateUser(ctx context.Context, user *domain.User, roleIDs []uint64, badgeIDs []uint64, meta domain.AuditMetadata) error {
	// Check if user already exists
	if _, err := u.userRepo.GetByEmail(ctx, user.Email); err == nil {
		return errors.New("a user with this email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	
	// Default score format if missing
	if user.ScoreFormat == nil || *user.ScoreFormat == "" {
		defaultFormat := "1-10"
		user.ScoreFormat = &defaultFormat
	}

	// Generate Unique Slug
	if user.Slug == nil || *user.Slug == "" {
		slug := u.generateUniqueUserSlug(ctx, user.Name)
		user.Slug = &slug
	}

	// Generate UUID
	if user.UUID == "" {
		user.UUID = uuid.New().String()
	}

	// Create User
	if err := u.userRepo.Create(ctx, user); err != nil {
		return err
	}

	// Attach roles
	if err := u.userRepo.UpdateRoles(ctx, user.ID, roleIDs); err != nil {
		return err
	}

	if err := u.userRepo.UpdateBadges(ctx, user.ID, badgeIDs); err != nil {
		return err
	}

	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "created", user.ID, "user", nil, user, &meta.URL, &meta.IPAddress, &meta.UserAgent)

	return nil
}

func (u *UserAdminUsecase) UpdateUser(ctx context.Context, user *domain.User, roleIDs []uint64, badgeIDs []uint64, meta domain.AuditMetadata) error {
	existing, err := u.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return err
	}

	existing.Name = user.Name
	existing.Email = user.Email

	if existing.Slug == nil || *existing.Slug == "" {
		slug := u.generateUniqueUserSlug(ctx, user.Name)
		existing.Slug = &slug
	}

	if err := u.userRepo.Update(ctx, existing); err != nil {
		return err
	}

	if err := u.userRepo.UpdateRoles(ctx, existing.ID, roleIDs); err != nil {
		return err
	}

	if err := u.userRepo.UpdateBadges(ctx, existing.ID, badgeIDs); err != nil {
		return err
	}

	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "updated", existing.ID, "user", existing, user, &meta.URL, &meta.IPAddress, &meta.UserAgent)

	return nil
}

func (u *UserAdminUsecase) DeleteUser(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	existing, _ := u.userRepo.GetByID(ctx, id)
	if err := u.userRepo.Delete(ctx, id); err != nil {
		return err
	}
	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "deleted", id, "user", existing, nil, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *UserAdminUsecase) ResetPassword(ctx context.Context, id uint64) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	var pw strings.Builder
	for i := 0; i < 10; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		pw.WriteByte(charset[n.Int64()])
	}
	rawPassword := pw.String()

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	if err := u.userRepo.UpdatePassword(ctx, id, string(hashedPassword)); err != nil {
		return "", err
	}

	return rawPassword, nil
}

func (u *UserAdminUsecase) generateUniqueUserSlug(ctx context.Context, name string) string {
	return utils.GenerateUniqueSlug(name, func(slug string) bool {
		existing, err := u.userRepo.GetBySlug(ctx, slug)
		return err == nil && existing != nil
	})
}

func (u *UserAdminUsecase) ResolveUserURLs(user *domain.User) {
	if user == nil || u.mediaService == nil {
		return
	}
	if user.Avatar != nil && *user.Avatar != "" {
		user.AvatarUrl = u.mediaService.Resolve(user.Avatar)
	}
	if user.Banner != nil && *user.Banner != "" {
		user.BannerUrl = u.mediaService.Resolve(user.Banner)
	}
}

func (u *UserAdminUsecase) ResolveUsersURLs(users []domain.User) {
	for i := range users {
		u.ResolveUserURLs(&users[i])
	}
}
