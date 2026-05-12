package admin

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
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
	badges = domain.FilterHighestBadges(badges)
	
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
	// 1. Load target and actor roles
	targetRoles, _ := u.userRepo.GetRolesByUserID(ctx, user.ID)
	actorRoles, err := u.userRepo.GetRolesByUserID(ctx, meta.ActorID)
	if err != nil {
		return err
	}

	// 2. Self-modification guard for roles (Optional, but safer for hierarchy)
	// If actor is not owner, they can only modify themselves if their weight > target's highest role weight?
	// Actually, let's keep it simple: Actor Max Weight >= Target Max Weight to UPDATE (except roles).
	// But let's stricter: Actor Weight >= Target Weight.
	targetMaxWeight := -1
	for _, r := range targetRoles {
		if r.Weight > targetMaxWeight {
			targetMaxWeight = r.Weight
		}
	}

	actorMaxWeight := -1
	for _, r := range actorRoles {
		if r.Weight > actorMaxWeight {
			actorMaxWeight = r.Weight
		}
	}

	// Rule: You cannot edit someone of equal or higher rank (unless you are editing yourself?)
	// Usually, admins can edit their own basic info, but not their own roles.
	if meta.ActorID != user.ID && actorMaxWeight <= targetMaxWeight {
		return domain.NewAppError(403, "Access denied. Your role level is insufficient to update this user.", nil)
	}

	// 3. Load existing user
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

	// 4. Role Update Guard: You cannot assign roles with weight > your max weight
	// However, for now let's just use the current permission gated logic.

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
	// 1. Prevent self-deletion
	if meta.ActorID == id {
		return domain.NewAppError(403, "You cannot delete your own account from the admin panel", nil)
	}

	// 2. Load target user and their roles
	targetUser, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	targetRoles, _ := u.userRepo.GetRolesByUserID(ctx, id)
	
	// Get target's max weight
	targetMaxWeight := -1
	for _, r := range targetRoles {
		if r.Weight > targetMaxWeight {
			targetMaxWeight = r.Weight
		}
	}

	// 3. Load actor and their roles
	actorRoles, err := u.userRepo.GetRolesByUserID(ctx, meta.ActorID)
	if err != nil {
		return err
	}

	// Get actor's max weight
	actorMaxWeight := -1
	for _, r := range actorRoles {
		if r.Weight > actorMaxWeight {
			actorMaxWeight = r.Weight
		}
	}

	// 4. Role Hierarchy Check: Actor must have strictly MORE weight than target
	// Exception: Owners (weight 100) are handled by the same logic, but we could add more guards if needed.
	if actorMaxWeight <= targetMaxWeight {
		return domain.NewAppError(403, "Access denied. Your role level is insufficient to delete this user.", nil)
	}

	// 5. Proceed with deletion
	if err := u.userRepo.Delete(ctx, id); err != nil {
		return err
	}

	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "deleted", id, "user", targetUser, nil, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *UserAdminUsecase) ResetPassword(ctx context.Context, id uint64, meta domain.AuditMetadata) (string, error) {
	// 1. Role Hierarchy Check
	targetRoles, _ := u.userRepo.GetRolesByUserID(ctx, id)
	actorRoles, err := u.userRepo.GetRolesByUserID(ctx, meta.ActorID)
	if err != nil {
		return "", err
	}

	targetMaxWeight := -1
	for _, r := range targetRoles {
		if r.Weight > targetMaxWeight {
			targetMaxWeight = r.Weight
		}
	}

	actorMaxWeight := -1
	for _, r := range actorRoles {
		if r.Weight > actorMaxWeight {
			actorMaxWeight = r.Weight
		}
	}

	// Actor must have weight >= Target for password reset? (Self-reset is handled elsewhere)
	if meta.ActorID != id && actorMaxWeight <= targetMaxWeight {
		return "", domain.NewAppError(403, "Access denied. Insufficient role to reset password for this user.", nil)
	}

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
	
	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "reset_password", id, "user", nil, nil, &meta.URL, &meta.IPAddress, &meta.UserAgent)

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

func (u *UserAdminUsecase) ProcessUserImages(ctx context.Context, progress chan<- string) error {
	sendProgress := func(msg string) {
		if progress != nil {
			select {
			case progress <- msg:
			default:
			}
		}
	}

	users, _, err := u.userRepo.GetUsers(ctx, 1, 10000, "")
	if err != nil {
		return err
	}

	sendProgress(fmt.Sprintf("Processing %d users for avatars and banners...", len(users)))

	for i, user := range users {
		// Process Avatar
		if user.Avatar != nil && *user.Avatar != "" {
			sendProgress(fmt.Sprintf("[%d/%d] Processing avatar for: %s", i+1, len(users), user.Name))
			file, err := u.mediaService.GetFile(ctx, *user.Avatar)
			if err == nil {
				newPath, _, err := u.mediaService.UploadWithResolutions(ctx, "users/avatars", user.ID, file, infrastructure.PresetSquare)
				file.Close()
				if err == nil {
					oldPath := *user.Avatar
					if err := u.userRepo.SetImage(ctx, user.ID, "avatar", newPath); err == nil {
						if oldPath != newPath && !strings.HasPrefix(oldPath, "http") {
							u.mediaService.DeleteMedia(ctx, oldPath)
						}
					}
				}
			}
		}

		// Process Banner
		if user.Banner != nil && *user.Banner != "" {
			sendProgress(fmt.Sprintf("[%d/%d] Processing banner for: %s", i+1, len(users), user.Name))
			file, err := u.mediaService.GetFile(ctx, *user.Banner)
			if err == nil {
				newPath, _, err := u.mediaService.UploadWithResolutions(ctx, "users/banners", user.ID, file, infrastructure.PresetLandscape)
				file.Close()
				if err == nil {
					oldPath := *user.Banner
					if err := u.userRepo.SetImage(ctx, user.ID, "banner", newPath); err == nil {
						if oldPath != newPath && !strings.HasPrefix(oldPath, "http") {
							u.mediaService.DeleteMedia(ctx, oldPath)
						}
					}
				}
			}
		}
	}

	sendProgress("Completed user image processing!")
	return nil
}
