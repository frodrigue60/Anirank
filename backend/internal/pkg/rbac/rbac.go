package rbac

import (
	"context"
	"sync"
	"time"

	"anirank/api/internal/domain"
)

// PermissionManager handles the caching of role-permission mappings.
type PermissionManager struct {
	repo       domain.UserRepository
	cache      map[string]map[string]bool // role_slug -> permission_slugs
	lastUpdate time.Time
	mu         sync.RWMutex
}

var (
	manager *PermissionManager
	once    sync.Once
)

// GetPermissionManager returns a singleton instance of the PermissionManager.
func GetPermissionManager(repo domain.UserRepository) *PermissionManager {
	once.Do(func() {
		manager = &PermissionManager{
			repo:  repo,
			cache: make(map[string]map[string]bool),
		}
		// Initial load (Best effort)
		_ = manager.Refresh(context.Background())
	})
	return manager
}

// Refresh reloads all permissions from the database.
func (m *PermissionManager) Refresh(ctx context.Context) error {
	roles, err := m.repo.GetRoles(ctx)
	if err != nil {
		return err
	}

	newCache := make(map[string]map[string]bool)
	for _, role := range roles {
		perms, err := m.repo.GetPermissionsByRoleID(ctx, role.ID)
		if err != nil {
			continue
		}
		rolePerms := make(map[string]bool)
		for _, p := range perms {
			rolePerms[p.Slug] = true
		}
		newCache[role.Slug] = rolePerms
	}

	m.mu.Lock()
	m.cache = newCache
	m.lastUpdate = time.Now()
	m.mu.Unlock()
	
	return nil
}

// HasPermission checks if a role has a specific permission.
func (m *PermissionManager) HasPermission(roleSlug, permSlug string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if perms, ok := m.cache[roleSlug]; ok {
		return perms[permSlug]
	}
	return false
}
