package postgres

import (
	"context"
	"fmt"
	"log"
	"strings"

	"anirank/api/internal/pkg/rbac"

	"github.com/jmoiron/sqlx"
)

// PermissionSeeder synchronizes RBAC permissions and role mappings
// from the Go-defined registry to the database on every startup.
type PermissionSeeder struct {
	db *sqlx.DB
}

func NewPermissionSeeder(db *sqlx.DB) *PermissionSeeder {
	return &PermissionSeeder{db: db}
}

// Seed orchestrates the full RBAC sync: permissions upsert, role linking, and cleanup.
func (s *PermissionSeeder) Seed(ctx context.Context) error {
	perms := rbac.BuildPermissionRegistry()
	matrix := rbac.BuildRoleMatrix(perms)

	// Step 1: Upsert all permissions
	synced, err := s.syncPermissions(ctx, perms)
	if err != nil {
		return fmt.Errorf("failed to sync permissions: %w", err)
	}

	// Step 2: Sync role-permission mappings (additive + cleanup)
	linked, err := s.syncRolePermissions(ctx, matrix)
	if err != nil {
		return fmt.Errorf("failed to sync role permissions: %w", err)
	}

	// Step 3: Cleanup orphaned permissions not in the registry
	removed, err := s.cleanupOrphanedPermissions(ctx, perms)
	if err != nil {
		log.Printf("⚠️  RBAC: Warning during orphan cleanup: %v", err)
		// Non-fatal: don't block startup
	}

	log.Printf("🔐 RBAC: Synced %d permissions, linked to roles (%s), removed %d orphans",
		synced, formatLinked(linked), removed)

	return nil
}

// syncPermissions upserts each PermissionDef into the permissions table.
func (s *PermissionSeeder) syncPermissions(ctx context.Context, perms []rbac.PermissionDef) (int, error) {
	query := `
		INSERT INTO permissions (name, slug, description, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (slug) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			updated_at = NOW()
	`

	count := 0
	for _, p := range perms {
		if _, err := s.db.ExecContext(ctx, query, p.Name, p.Slug, p.Description); err != nil {
			return count, fmt.Errorf("failed to upsert permission %q: %w", p.Slug, err)
		}
		count++
	}
	return count, nil
}

// syncRolePermissions performs a full sync of role-permission mappings.
// For each role in the matrix, it deletes all existing mappings and re-inserts the correct ones.
func (s *PermissionSeeder) syncRolePermissions(ctx context.Context, matrix rbac.RolePermissionMatrix) (map[string]int, error) {
	linked := make(map[string]int)

	for roleSlug, permSlugs := range matrix {
		if len(permSlugs) == 0 {
			continue
		}

		// Get role ID
		var roleID uint64
		err := s.db.GetContext(ctx, &roleID, "SELECT id FROM roles WHERE slug = $1", roleSlug)
		if err != nil {
			log.Printf("⚠️  RBAC: Role %q not found in DB, skipping", roleSlug)
			continue
		}

		// Start transaction for atomic role sync
		tx, err := s.db.BeginTxx(ctx, nil)
		if err != nil {
			return linked, fmt.Errorf("failed to begin tx for role %q: %w", roleSlug, err)
		}

		// Delete all current mappings for this role
		if _, err := tx.ExecContext(ctx, "DELETE FROM role_permissions WHERE role_id = $1", roleID); err != nil {
			tx.Rollback()
			return linked, fmt.Errorf("failed to clear role_permissions for %q: %w", roleSlug, err)
		}

		// Re-insert from matrix
		count := 0
		for _, permSlug := range permSlugs {
			result, err := tx.ExecContext(ctx, `
				INSERT INTO role_permissions (role_id, permission_id)
				SELECT $1, id FROM permissions WHERE slug = $2
				ON CONFLICT DO NOTHING
			`, roleID, permSlug)
			if err != nil {
				tx.Rollback()
				return linked, fmt.Errorf("failed to link %q -> %q: %w", roleSlug, permSlug, err)
			}
			rows, _ := result.RowsAffected()
			count += int(rows)
		}

		if err := tx.Commit(); err != nil {
			return linked, fmt.Errorf("failed to commit role %q: %w", roleSlug, err)
		}

		linked[roleSlug] = count
	}

	return linked, nil
}

// cleanupOrphanedPermissions removes permissions from the DB that are
// no longer defined in the registry. This prevents stale data accumulation.
func (s *PermissionSeeder) cleanupOrphanedPermissions(ctx context.Context, perms []rbac.PermissionDef) (int, error) {
	if len(perms) == 0 {
		return 0, nil
	}

	// Build the list of valid slugs
	validSlugs := make([]string, 0, len(perms))
	for _, p := range perms {
		validSlugs = append(validSlugs, p.Slug)
	}

	// Use sqlx.In for safe IN clause
	query, args, err := sqlx.In("DELETE FROM permissions WHERE slug NOT IN (?)", validSlugs)
	if err != nil {
		return 0, fmt.Errorf("failed to build cleanup query: %w", err)
	}

	query = s.db.Rebind(query)
	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to delete orphaned permissions: %w", err)
	}

	rows, _ := result.RowsAffected()
	return int(rows), nil
}

// formatLinked creates a human-readable summary of role linking counts.
func formatLinked(linked map[string]int) string {
	parts := make([]string, 0, len(linked))
	for role, count := range linked {
		parts = append(parts, fmt.Sprintf("%s: %d", role, count))
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ", ")
}
