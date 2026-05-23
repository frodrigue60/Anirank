package postgres

import (
	"context"
	"testing"

	"anirank/api/internal/pkg/rbac"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

// TestPermissionSeeder_SyncPermissions_ExecutesUpsertForEachPermission verifies
// that the seeder executes an INSERT...ON CONFLICT for every permission in the registry.
func TestPermissionSeeder_SyncPermissions_ExecutesUpsertForEachPermission(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	seeder := NewPermissionSeeder(sqlxDB)

	perms := rbac.BuildPermissionRegistry()

	// Expect one INSERT per permission (upsert via ON CONFLICT)
	for _, p := range perms {
		mock.ExpectExec("INSERT INTO permissions").
			WithArgs(p.Name, p.Slug, p.Description).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	count, err := seeder.syncPermissions(context.Background(), perms)

	assert.NoError(t, err)
	assert.Equal(t, len(perms), count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestPermissionSeeder_SyncRolePermissions_ExecutesLinkingPerRole verifies
// that for each role in the matrix, the seeder:
//   1. Queries the role ID by slug
//   2. Deletes existing role_permissions
//   3. Inserts new role_permissions for each permission slug
func TestPermissionSeeder_SyncRolePermissions_ExecutesLinkingPerRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	seeder := NewPermissionSeeder(sqlxDB)

	// Use a minimal matrix to keep expectations manageable
	matrix := rbac.RolePermissionMatrix{
		"admin": {"anime.create", "anime.edit"},
	}

	// Expect: SELECT role id by slug
	mock.ExpectQuery("SELECT id FROM roles WHERE slug").
		WithArgs("admin").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Expect: BEGIN transaction
	mock.ExpectBegin()

	// Expect: DELETE existing role_permissions
	mock.ExpectExec("DELETE FROM role_permissions WHERE role_id").
		WithArgs(uint64(1)).
		WillReturnResult(sqlmock.NewResult(0, 5)) // pretend 5 were deleted

	// Expect: INSERT for each permission slug
	mock.ExpectExec("INSERT INTO role_permissions").
		WithArgs(uint64(1), "anime.create").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("INSERT INTO role_permissions").
		WithArgs(uint64(1), "anime.edit").
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Expect: COMMIT
	mock.ExpectCommit()

	linked, err := seeder.syncRolePermissions(context.Background(), matrix)

	assert.NoError(t, err)
	assert.Equal(t, 2, linked["admin"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestPermissionSeeder_SyncRolePermissions_SkipsMissingRole verifies
// that when a role doesn't exist in the DB, it's skipped without error.
func TestPermissionSeeder_SyncRolePermissions_SkipsMissingRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	seeder := NewPermissionSeeder(sqlxDB)

	matrix := rbac.RolePermissionMatrix{
		"nonexistent_role": {"anime.create"},
	}

	// Role query returns no rows → sql.ErrNoRows
	mock.ExpectQuery("SELECT id FROM roles WHERE slug").
		WithArgs("nonexistent_role").
		WillReturnRows(sqlmock.NewRows([]string{"id"})) // empty result set

	linked, err := seeder.syncRolePermissions(context.Background(), matrix)

	assert.NoError(t, err)
	assert.Equal(t, 0, linked["nonexistent_role"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestPermissionSeeder_CleanupOrphanedPermissions_DeletesStaleEntries verifies
// that permissions NOT in the registry are deleted from the DB.
func TestPermissionSeeder_CleanupOrphanedPermissions_DeletesStaleEntries(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	seeder := NewPermissionSeeder(sqlxDB)

	// Small permission set — the DELETE query should use these slugs in the NOT IN clause
	perms := []rbac.PermissionDef{
		{Name: "Create Anime", Slug: "anime.create", Description: "test"},
		{Name: "Edit Anime", Slug: "anime.edit", Description: "test"},
	}

	// Expect DELETE with NOT IN containing our valid slugs
	mock.ExpectExec("DELETE FROM permissions WHERE slug NOT IN").
		WithArgs("anime.create", "anime.edit").
		WillReturnResult(sqlmock.NewResult(0, 3)) // pretend 3 orphans deleted

	removed, err := seeder.cleanupOrphanedPermissions(context.Background(), perms)

	assert.NoError(t, err)
	assert.Equal(t, 3, removed)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestPermissionSeeder_Seed_FullFlow verifies the complete Seed() orchestration
// with the full permission registry.
func TestPermissionSeeder_Seed_FullFlow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// Go maps iterate in random order, so we must allow out-of-order matching
	mock.MatchExpectationsInOrder(false)

	sqlxDB := sqlx.NewDb(db, "postgres")
	seeder := NewPermissionSeeder(sqlxDB)

	perms := rbac.BuildPermissionRegistry()
	matrix := rbac.BuildRoleMatrix(perms)

	// --- Step 1: Expect INSERT for each permission ---
	for _, p := range perms {
		mock.ExpectExec("INSERT INTO permissions").
			WithArgs(p.Name, p.Slug, p.Description).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	// --- Step 2: Expect role linking for each role in the matrix ---
	roleIDs := map[string]uint64{
		"admin":   1,
		"editor":  2,
		"creator": 3,
	}

	for roleSlug, permSlugs := range matrix {
		roleID := roleIDs[roleSlug]

		// SELECT role
		mock.ExpectQuery("SELECT id FROM roles WHERE slug").
			WithArgs(roleSlug).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(roleID))

		// BEGIN tx
		mock.ExpectBegin()

		// DELETE existing
		mock.ExpectExec("DELETE FROM role_permissions WHERE role_id").
			WithArgs(roleID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		// INSERT for each perm
		for _, permSlug := range permSlugs {
			mock.ExpectExec("INSERT INTO role_permissions").
				WithArgs(roleID, permSlug).
				WillReturnResult(sqlmock.NewResult(0, 1))
		}

		// COMMIT
		mock.ExpectCommit()
	}

	// --- Step 3: Expect cleanup of orphaned permissions ---
	// Use a flexible expectation since sqlx.In generates dynamic args
	mock.ExpectExec("DELETE FROM permissions WHERE slug NOT IN").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = seeder.Seed(context.Background())

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
