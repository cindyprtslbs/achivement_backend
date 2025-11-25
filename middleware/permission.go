package middleware

import (
	"achievement_backend/app/repository"
	"github.com/gofiber/fiber/v2"
)

type RBACMiddleware struct {
	rolePermissionRepo repository.RolePermissionRepository
}

func NewRBACMiddleware(rp repository.RolePermissionRepository) *RBACMiddleware {
	return &RBACMiddleware{
		rolePermissionRepo: rp,
	}
}

func (m *RBACMiddleware) PermissionRequired(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {

		roleID := c.Locals("role_id")
		if roleID == nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "unauthorized: missing role",
			})
		}

		roleIDStr := roleID.(string)

		// FR-002 Step 3:
		// Load permissions dari database/cache
		perms, err := m.rolePermissionRepo.GetPermissionsByRole(roleIDStr)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "failed to load permissions",
			})
		}

		// FR-002 Step 4:
		// Check apakah user memiliki permission yg diminta
		for _, p := range perms {
			full := p.Resource + ":" + p.Action
			if full == permission {
				return c.Next()
			}
		}

		// FR-002 Step 5:
		// Deny request
		return c.Status(403).JSON(fiber.Map{
			"error": "forbidden: missing permission `" + permission + "`",
		})
	}
}
