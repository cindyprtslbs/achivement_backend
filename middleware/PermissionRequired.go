package middleware

import (
	"github.com/gofiber/fiber/v2"
)

func PermissionRequired(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {

		perms := c.Locals("permissions")
		if perms == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized: permissions not found",
			})
		}

		permList, ok := perms.([]string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "forbidden: invalid permissions format",
			})
		}

		for _, p := range permList {
			if p == permission {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "forbidden: missing permission `" + permission + "`",
		})
	}
}

