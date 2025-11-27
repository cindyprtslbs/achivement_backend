package middleware

import (
	"achievement_backend/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {

		token := c.Get("Authorization")
		if token == "" {
			return c.Status(401).JSON(fiber.Map{"error": "missing token"})
		}

		parts := strings.Split(token, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(401).JSON(fiber.Map{"error": "invalid token format"})
		}

		rawToken := parts[1]

		// CEK BLACKLIST
		if utils.IsBlacklisted(rawToken) {
			return c.Status(401).JSON(fiber.Map{
				"error": "token blacklisted (logged out)",
			})
		}

		// VALIDATE TOKEN
		claims, err := utils.ValidateToken(rawToken)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "invalid token"})
		}

		// SET CONTEXT
		c.Locals("raw_token", rawToken)
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("role_name", claims.RoleName)
		c.Locals("permissions", claims.Permissions)

		return c.Next()
	}
}
