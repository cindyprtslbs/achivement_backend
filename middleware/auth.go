package middleware

import (
	"strings"

	"achievement_backend/utils"

	"github.com/gofiber/fiber/v2"
)

func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{"error": "token required"})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(401).JSON(fiber.Map{"error": "invalid token format"})
		}

		claims, err := utils.ValidateToken(parts[1])
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "invalid or expired token"})
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("role_id", claims.RoleID)
		c.Locals("permissions", claims.Permissions)

		return c.Next()
	}
}

func RequirePermission(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		perms := c.Locals("permissions").([]string)

		for _, p := range perms {
			if p == permission {
				return c.Next()
			}
		}

		return c.Status(403).JSON(fiber.Map{"error": "forbidden"})
	}
}
