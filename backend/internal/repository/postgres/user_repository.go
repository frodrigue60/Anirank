package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"anirank/api/internal/domain"

	"github.com/jmoiron/sqlx"
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) domain.UserRepository {
	return &userRepository{db: db.Unsafe()}
}

func (r *userRepository) GetByID(ctx context.Context, id uint64) (*domain.User, error) {
	var user domain.User
	query := `
		SELECT u.*, sf.slug AS score_format,
		       (SELECT COUNT(*) FROM follows WHERE followed_id = u.id) as followers_count,
		       (SELECT COUNT(*) FROM follows WHERE follower_id = u.id) as following_count
		FROM users u
		LEFT JOIN score_formats sf ON u.score_format_id = sf.id
		WHERE u.id = $1
	`
	err := r.db.GetContext(ctx, &user, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err == nil {
		ids, _ := r.GetSocialIdentitiesByUserID(ctx, user.ID)
		user.SocialIdentities = ids
	}
	return &user, err
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	query := `
		SELECT u.*, sf.slug AS score_format
		FROM users u
		LEFT JOIN score_formats sf ON u.score_format_id = sf.id
		WHERE u.email = $1
	`
	err := r.db.GetContext(ctx, &user, query, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err == nil {
		ids, _ := r.GetSocialIdentitiesByUserID(ctx, user.ID)
		user.SocialIdentities = ids
	}
	return &user, err
}

func (r *userRepository) GetByGoogleID(ctx context.Context, googleID string) (*domain.User, error) {
	var user domain.User
	query := `
		SELECT u.*, sf.slug AS score_format
		FROM users u
		JOIN user_social_identities usi ON u.id = usi.user_id
		LEFT JOIN score_formats sf ON u.score_format_id = sf.id
		WHERE usi.provider = 'google' AND usi.provider_id = $1
	`
	err := r.db.GetContext(ctx, &user, query, googleID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err == nil {
		ids, _ := r.GetSocialIdentitiesByUserID(ctx, user.ID)
		user.SocialIdentities = ids
	}
	return &user, err
}

func (r *userRepository) GetByAnilistID(ctx context.Context, anilistID uint64) (*domain.User, error) {
	var user domain.User
	query := `
		SELECT u.*, sf.slug AS score_format
		FROM users u
		JOIN user_social_identities usi ON u.id = usi.user_id
		LEFT JOIN score_formats sf ON u.score_format_id = sf.id
		WHERE usi.provider = 'anilist' AND usi.provider_id = $1
	`
	err := r.db.GetContext(ctx, &user, query, fmt.Sprintf("%d", anilistID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err == nil {
		ids, _ := r.GetSocialIdentitiesByUserID(ctx, user.ID)
		user.SocialIdentities = ids
	}
	return &user, err
}

func (r *userRepository) GetBySlug(ctx context.Context, slug string) (*domain.User, error) {
	var user domain.User
	query := `
		SELECT u.*, sf.slug AS score_format,
		       (SELECT COUNT(*) FROM follows WHERE followed_id = u.id) as followers_count,
		       (SELECT COUNT(*) FROM follows WHERE follower_id = u.id) as following_count,
		       (SELECT COUNT(*) FROM song_ratings WHERE user_id = u.id) as ratings_count
		FROM users u
		LEFT JOIN score_formats sf ON u.score_format_id = sf.id
		WHERE u.slug = $1
	`
	err := r.db.GetContext(ctx, &user, query, slug)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err == nil {
		ids, _ := r.GetSocialIdentitiesByUserID(ctx, user.ID)
		user.SocialIdentities = ids
	}
	return &user, err
}


func (r *userRepository) GetByUUID(ctx context.Context, uuid string) (*domain.User, error) {
	var user domain.User
	query := `
		SELECT u.*, sf.slug AS score_format,
		       (SELECT COUNT(*) FROM follows WHERE followed_id = u.id) as followers_count,
		       (SELECT COUNT(*) FROM follows WHERE follower_id = u.id) as following_count,
		       (SELECT COUNT(*) FROM song_ratings WHERE user_id = u.id) as ratings_count
		FROM users u
		LEFT JOIN score_formats sf ON u.score_format_id = sf.id
		WHERE u.uuid = $1
	`
	err := r.db.GetContext(ctx, &user, query, uuid)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err == nil {
		ids, _ := r.GetSocialIdentitiesByUserID(ctx, user.ID)
		user.SocialIdentities = ids
	}
	return &user, err
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (uuid, name, slug, email, password, score_format_id, avatar, banner, about, profile_color, created_at, updated_at)
		VALUES (:uuid, :name, :slug, :email, :password, (SELECT id FROM score_formats WHERE slug = :score_format), :avatar, :banner, :about, :profile_color, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`
	rows, err := r.db.NamedQueryContext(ctx, query, user)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.Scan(&user.ID)
	}
	return nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	query := `UPDATE users SET name=:name, slug=:slug, email=:email, score_format_id=(SELECT id FROM score_formats WHERE slug = :score_format), avatar=:avatar, banner=:banner, about=:about, profile_color=:profile_color, updated_at=CURRENT_TIMESTAMP WHERE id=:id`
	_, err := r.db.NamedExecContext(ctx, query, user)
	return err
}

func (r *userRepository) UpdatePassword(ctx context.Context, userID uint64, hashedPassword string) error {
	query := `UPDATE users SET password = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, hashedPassword, userID)
	return err
}

func (r *userRepository) SetImage(ctx context.Context, userID uint64, imageType, imagePath string) error {
	column := "avatar"
	if imageType == "banner" {
		column = "banner"
	}

	query := "UPDATE users SET " + column + " = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2"
	_, err := r.db.ExecContext(ctx, query, imagePath, userID)
	return err
}

func (r *userRepository) UpdateScoreFormat(ctx context.Context, userID uint64, format string) error {
	query := `
		UPDATE users 
		SET score_format_id = (SELECT id FROM score_formats WHERE slug = $1 LIMIT 1), 
		updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, format, userID)
	return err
}

func (r *userRepository) GetRolesByUserID(ctx context.Context, userID uint64) ([]domain.Role, error) {
	var roles []domain.Role
	query := `
		SELECT r.id, r.name, r.slug, r.weight 
		FROM roles r 
		JOIN role_user ru ON r.id = ru.role_id 
		WHERE ru.user_id = $1
	`
	err := r.db.SelectContext(ctx, &roles, query, userID)
	return roles, err
}

func (r *userRepository) GetPermissionsByRoleID(ctx context.Context, roleID uint64) ([]domain.Permission, error) {
	var permissions []domain.Permission
	query := `
		SELECT p.id, p.name, p.slug, p.description, p.created_at, p.updated_at
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
	`
	err := r.db.SelectContext(ctx, &permissions, query, roleID)
	return permissions, err
}

func (r *userRepository) GetPermissionsByUserID(ctx context.Context, userID uint64) ([]domain.Permission, error) {
	var permissions []domain.Permission
	query := `
		SELECT DISTINCT p.id, p.name, p.slug, p.description, p.created_at, p.updated_at
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN role_user ru ON rp.role_id = ru.role_id
		WHERE ru.user_id = $1
	`
	err := r.db.SelectContext(ctx, &permissions, query, userID)
	return permissions, err
}

func (r *userRepository) GetAllPermissions(ctx context.Context) ([]domain.Permission, error) {
	var permissions []domain.Permission
	query := "SELECT id, name, slug, description, created_at, updated_at FROM permissions ORDER BY slug ASC"
	err := r.db.SelectContext(ctx, &permissions, query)
	return permissions, err
}

func (r *userRepository) GetRoles(ctx context.Context) ([]domain.Role, error) {
	var roles []domain.Role
	query := "SELECT id, name, slug, weight, description, created_at FROM roles ORDER BY weight DESC"
	err := r.db.SelectContext(ctx, &roles, query)
	return roles, err
}

func (r *userRepository) GetBadgesByUserID(ctx context.Context, userID uint64) ([]domain.Badge, error) {
	var badges []domain.Badge
	query := `
		SELECT b.id, b.uuid, b.name, b.description, b.icon, b.is_active 
		FROM badges b 
		JOIN badge_user bu ON b.id = bu.badge_id 
		WHERE bu.user_id = $1
	`
	err := r.db.SelectContext(ctx, &badges, query, userID)
	return badges, err
}

// GetBadgesByUserIDs fetches badges for multiple users in a single query,
// returning a map[userID][]Badge for O(1) lookup — prevents N+1 in comment enrichment.
func (r *userRepository) GetBadgesByUserIDs(ctx context.Context, userIDs []uint64) (map[uint64][]domain.Badge, error) {
	if len(userIDs) == 0 {
		return map[uint64][]domain.Badge{}, nil
	}

	type badgeWithUserID struct {
		domain.Badge
		UserID uint64 `db:"user_id"`
	}

	var rows []badgeWithUserID
	query, args, err := sqlx.In(`
		SELECT bu.user_id, b.id, b.uuid, b.name, b.description, b.icon, b.is_active
		FROM badges b
		JOIN badge_user bu ON b.id = bu.badge_id
		WHERE bu.user_id IN (?)
	`, userIDs)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, err
	}

	result := make(map[uint64][]domain.Badge, len(userIDs))
	for _, row := range rows {
		result[row.UserID] = append(result[row.UserID], row.Badge)
	}
	return result, nil
}

func (r *userRepository) GetUsers(ctx context.Context, page, limit int, search string) ([]domain.User, int, error) {
	var users []domain.User
	var total int

	offset := (page - 1) * limit
	query := `
		SELECT u.*, sf.slug AS score_format
		FROM users u
		LEFT JOIN score_formats sf ON u.score_format_id = sf.id
	`
	countQuery := `SELECT COUNT(*) FROM users u`
	var args []interface{}

	if search != "" {
		condition := " WHERE u.name ILIKE $1 OR u.email ILIKE $2"
		query += condition
		countQuery += condition
		searchParam := "%" + search + "%"
		args = append(args, searchParam, searchParam)
	}

	argCount := len(args)
	query += fmt.Sprintf(" ORDER BY u.created_at DESC LIMIT $%d OFFSET $%d", argCount+1, argCount+2)

	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	if err := r.db.SelectContext(ctx, &users, query, args...); err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepository) Delete(ctx context.Context, id uint64) error {
	query := "DELETE FROM users WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *userRepository) UpdateRoles(ctx context.Context, userID uint64, roleIDs []uint64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Delete existing roles
	_, err = tx.ExecContext(ctx, "DELETE FROM role_user WHERE user_id = $1", userID)
	if err != nil {
		return err
	}

	// 2. Insert new roles
	if len(roleIDs) > 0 {
		query := "INSERT INTO role_user (user_id, role_id) VALUES ($1, $2)"
		stmt, err := tx.PrepareContext(ctx, query)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, roleID := range roleIDs {
			if _, err := stmt.ExecContext(ctx, userID, roleID); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *userRepository) UpdateBadges(ctx context.Context, userID uint64, badgeIDs []uint64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Delete existing badges
	_, err = tx.ExecContext(ctx, "DELETE FROM badge_user WHERE user_id = $1", userID)
	if err != nil {
		return err
	}

	// 2. Insert new badges
	if len(badgeIDs) > 0 {
		query := "INSERT INTO badge_user (user_id, badge_id) VALUES ($1, $2)"
		stmt, err := tx.PrepareContext(ctx, query)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, badgeID := range badgeIDs {
			if _, err := stmt.ExecContext(ctx, userID, badgeID); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *userRepository) UpdateRolePermissions(ctx context.Context, roleID uint64, permissionIDs []uint64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Delete existing permissions for the role
	_, err = tx.ExecContext(ctx, "DELETE FROM role_permissions WHERE role_id = $1", roleID)
	if err != nil {
		return err
	}

	// 2. Insert new permissions
	if len(permissionIDs) > 0 {
		query := "INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)"
		stmt, err := tx.PrepareContext(ctx, query)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, permID := range permissionIDs {
			if _, err := stmt.ExecContext(ctx, roleID, permID); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *userRepository) Search(ctx context.Context, term string, limit int) ([]domain.User, error) {
	var users []domain.User
	query := `
		SELECT id, uuid, name, slug, avatar 
		FROM users 
		WHERE name ILIKE $1 
		ORDER BY created_at DESC
		LIMIT $2
	`
	err := r.db.SelectContext(ctx, &users, query, term, limit)
	if users == nil {
		users = []domain.User{}
	}
	return users, err
}

func (r *userRepository) Follow(ctx context.Context, followerID, followedID uint64) error {
	query := "INSERT INTO follows (follower_id, followed_id, created_at) VALUES ($1, $2, CURRENT_TIMESTAMP)"
	_, err := r.db.ExecContext(ctx, query, followerID, followedID)
	return err
}

func (r *userRepository) Unfollow(ctx context.Context, followerID, followedID uint64) error {
	query := "DELETE FROM follows WHERE follower_id = $1 AND followed_id = $2"
	_, err := r.db.ExecContext(ctx, query, followerID, followedID)
	return err
}

func (r *userRepository) IsFollowing(ctx context.Context, followerID, followedID uint64) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND followed_id = $2)"
	err := r.db.GetContext(ctx, &exists, query, followerID, followedID)
	return exists, err
}

func (r *userRepository) GetFollowersCount(ctx context.Context, userID uint64) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM follows WHERE followed_id = $1"
	err := r.db.GetContext(ctx, &count, query, userID)
	return count, err
}

func (r *userRepository) GetFollowingCount(ctx context.Context, userID uint64) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM follows WHERE follower_id = $1"
	err := r.db.GetContext(ctx, &count, query, userID)
	return count, err
}

func (r *userRepository) GetFollowers(ctx context.Context, userID uint64, limit, offset int) ([]domain.User, error) {
	var users []domain.User
	query := `
		SELECT u.id, u.uuid, u.name, u.slug, u.avatar, u.banner
		FROM users u
		JOIN follows f ON u.id = f.follower_id
		WHERE f.followed_id = $1
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
	`
	err := r.db.SelectContext(ctx, &users, query, userID, limit, offset)
	if users == nil {
		users = []domain.User{}
	}
	return users, err
}

func (r *userRepository) GetFollowing(ctx context.Context, userID uint64, limit, offset int) ([]domain.User, error) {
	var users []domain.User
	query := `
		SELECT u.id, u.uuid, u.name, u.slug, u.avatar, u.banner
		FROM users u
		JOIN follows f ON u.id = f.followed_id
		WHERE f.follower_id = $1
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
	`
	err := r.db.SelectContext(ctx, &users, query, userID, limit, offset)
	if users == nil {
		users = []domain.User{}
	}
	return users, err
}

func (r *userRepository) GetRanking(ctx context.Context, sortBy string, limit, offset int) ([]domain.RankingUser, int, error) {
	var users []domain.RankingUser
	var total int

	// Base query with counts
	baseQuery := `
		SELECT u.*,
		       (SELECT COUNT(*) FROM song_ratings WHERE user_id = u.id) as ratings_count,
		       (SELECT COUNT(*) FROM comments WHERE user_id = u.id) as comments_count
		FROM users u
	`

	// Ordering
	orderBy := " ORDER BY u.level DESC, u.xp DESC "
	if sortBy == "ratings" {
		orderBy = " ORDER BY ratings_count DESC, u.level DESC, u.xp DESC "
	} else if sortBy == "comments" {
		orderBy = " ORDER BY comments_count DESC, u.level DESC, u.xp DESC "
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM users"
	if err := r.db.GetContext(ctx, &total, countQuery); err != nil {
		return nil, 0, err
	}

	// Final query with pagination
	query := baseQuery + orderBy + " LIMIT $1 OFFSET $2"
	err := r.db.SelectContext(ctx, &users, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepository) GetMany(ctx context.Context, ids []uint64) ([]domain.User, error) {
	if len(ids) == 0 {
		return []domain.User{}, nil
	}

	var users []domain.User
	query, args, err := sqlx.In("SELECT * FROM users WHERE id IN (?)", ids)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)
	err = r.db.SelectContext(ctx, &users, query, args...)
	if users == nil {
		users = []domain.User{}
	}
	return users, err
}

func (r *userRepository) GetSocialIdentity(ctx context.Context, provider, providerID string) (*domain.UserSocialIdentity, error) {
	var identity domain.UserSocialIdentity
	query := `SELECT * FROM user_social_identities WHERE provider = $1 AND provider_id = $2`
	err := r.db.GetContext(ctx, &identity, query, provider, providerID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &identity, err
}

func (r *userRepository) GetSocialIdentitiesByUserID(ctx context.Context, userID uint64) ([]domain.UserSocialIdentity, error) {
	var identities []domain.UserSocialIdentity
	query := `SELECT * FROM user_social_identities WHERE user_id = $1`
	err := r.db.SelectContext(ctx, &identities, query, userID)
	if identities == nil {
		identities = []domain.UserSocialIdentity{}
	}
	return identities, err
}

func (r *userRepository) SaveSocialIdentity(ctx context.Context, identity *domain.UserSocialIdentity) error {
	query := `
		INSERT INTO user_social_identities (user_id, provider, provider_id, provider_username, access_token, refresh_token, expires_at, created_at, updated_at)
		VALUES (:user_id, :provider, :provider_id, :provider_username, :access_token, :refresh_token, :expires_at, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (provider, provider_id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			provider_username = EXCLUDED.provider_username,
			access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			expires_at = EXCLUDED.expires_at,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.NamedExecContext(ctx, query, identity)
	return err
}

func (r *userRepository) DeleteSocialIdentity(ctx context.Context, userID uint64, provider string) error {
	query := `DELETE FROM user_social_identities WHERE user_id = $1 AND provider = $2`
	_, err := r.db.ExecContext(ctx, query, userID, provider)
	return err
}

