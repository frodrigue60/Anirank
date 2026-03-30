package admin

import (
	"context"
	"anirank/api/internal/domain"
	"anirank/api/internal/pkg/rbac"
)

type PermissionUsecase struct {
	userRepo domain.UserRepository
}

func NewPermissionUsecase(ur domain.UserRepository) *PermissionUsecase {
	return &PermissionUsecase{userRepo: ur}
}

func (u *PermissionUsecase) GetAllPermissions(ctx context.Context) ([]domain.Permission, error) {
	return u.userRepo.GetAllPermissions(ctx)
}

func (u *PermissionUsecase) GetRolePermissions(ctx context.Context, roleID uint64) ([]domain.Permission, error) {
	return u.userRepo.GetPermissionsByRoleID(ctx, roleID)
}

func (u *PermissionUsecase) UpdateRolePermissions(ctx context.Context, roleID uint64, permissionIDs []uint64) error {
	err := u.userRepo.UpdateRolePermissions(ctx, roleID, permissionIDs)
	if err != nil {
		return err
	}
	
	// Refresh the RBAC cache so changes take effect immediately
	pm := rbac.GetPermissionManager(u.userRepo)
	return pm.Refresh(ctx)
}

func (u *PermissionUsecase) GetRoles(ctx context.Context) ([]domain.Role, error) {
	return u.userRepo.GetRoles(ctx)
}
