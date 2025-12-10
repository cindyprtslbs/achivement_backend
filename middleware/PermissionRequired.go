package middleware

import (
	"github.com/gofiber/fiber/v2"
)

func PermissionRequired(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {

		perms := c.Locals("permissions")
		if perms == nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "unauthorized: maaf, kamu nggak punya permission apapun",
			})
		}

		permList := perms.([]string)

		// cek apakah permission yg diminta ada dalam token
		for _, p := range permList {
			if p == permission {
				return c.Next()
			}
		}

		return c.Status(403).JSON(fiber.Map{
			"error": "forbidden:  maaf, kamu nggak bisa akses `" + permission + "`",
		})
	}
}
