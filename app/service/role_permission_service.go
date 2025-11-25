package service

import (
	"github.com/gofiber/fiber/v2"
	"achievement_backend/app/repository"
)

type RolePermissionService struct {
	repo repository.RolePermissionRepository
}

func NewRolePermissionService(r repository.RolePermissionRepository) *RolePermissionService {
	return &RolePermissionService{repo: r}
}

//
// ===================== ASSIGN PERMISSION TO ROLE =====================
//
func (s *RolePermissionService) Assign(c *fiber.Ctx) error {
	var req struct {
		RoleID       string `json:"role_id"`
		PermissionID string `json:"permission_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.RoleID == "" || req.PermissionID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "role_id and permission_id are required",
		})
	}

	err := s.repo.AssignPermission(req.RoleID, req.PermissionID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to assign permission",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}

//
// ===================== REMOVE PERMISSION FROM ROLE =====================
//
func (s *RolePermissionService) Remove(c *fiber.Ctx) error {
	var req struct {
		RoleID       string `json:"role_id"`
		PermissionID string `json:"permission_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.RoleID == "" || req.PermissionID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "role_id and permission_id are required",
		})
	}

	err := s.repo.RemovePermission(req.RoleID, req.PermissionID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "role or permission not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}

//
// ===================== GET PERMISSIONS BY ROLE =====================
//
func (s *RolePermissionService) GetPermissionsByRole(c *fiber.Ctx) error {
	roleID := c.Params("role_id")

	if roleID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "role_id required",
		})
	}

	perms, err := s.repo.GetPermissionsByRole(roleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to fetch permissions",
		})
	}

	return c.JSON(perms)
}

//
// ===================== GET ROLES BY PERMISSION =====================
//
func (s *RolePermissionService) GetRolesByPermission(c *fiber.Ctx) error {
	permissionID := c.Params("permission_id")

	if permissionID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "permission_id required",
		})
	}

	roles, err := s.repo.GetRolesByPermission(permissionID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to fetch roles",
		})
	}

	return c.JSON(roles)
}
