package v1

import (
	"anirank/api/internal/domain"
	"anirank/api/internal/usecase/admin"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

type PermissionHandler struct {
	usecase *admin.PermissionUsecase
}

func NewPermissionHandler(u *admin.PermissionUsecase) *PermissionHandler {
	return &PermissionHandler{usecase: u}
}

// GetAllPermissions returns all available system permissions
func (h *PermissionHandler) GetAllPermissions(c *fiber.Ctx) error {
	perms, err := h.usecase.GetAllPermissions(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(domain.Response{
		Success: true,
		Data:    perms,
	})
}

// GetRoles returns all roles with their assigned permissions
func (h *PermissionHandler) GetRoles(c *fiber.Ctx) error {
	roles, err := h.usecase.GetRoles(c.Context())
	if err != nil {
		return err
	}
	
	// Load permissions for each role
	for i := range roles {
		perms, _ := h.usecase.GetRolePermissions(c.Context(), roles[i].ID)
		roles[i].Permissions = perms
	}
	
	return c.JSON(domain.Response{
		Success: true,
		Data:    roles,
	})
}

// UpdateRolePermissions updates the permissions for a specific role
func (h *PermissionHandler) UpdateRolePermissions(c *fiber.Ctx) error {
	roleIDStr := c.Params("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid role ID", nil)
	}

	var req struct {
		PermissionIDs []uint64 `json:"permission_ids"`
	}

	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid request body", nil)
	}

	if err := h.usecase.UpdateRolePermissions(c.Context(), roleID, req.PermissionIDs); err != nil {
		return err
	}

	return c.JSON(domain.Response{
		Success: true,
		Message: "Role permissions updated successfully",
	})
}
