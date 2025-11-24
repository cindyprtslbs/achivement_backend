package middleware

// import (
// 	"crud-app/utils"
// 	"strings"

// 	"github.com/gofiber/fiber/v2"
// )

// Middleware untuk memerlukan login
// func AuthRequired() fiber.Handler {
// 	return func(c *fiber.Ctx) error {
// 		authHeader := c.Get("Authorization")
// 		if authHeader == "" {
// 			return c.Status(401).JSON(fiber.Map{
// 				"error": "Token akses diperlukan",
// 			})
// 		}

// 		tokenParts := strings.Split(authHeader, " ")
// 		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
// 			return c.Status(401).JSON(fiber.Map{
// 				"error": "Format token tidak valid",
// 			})
// 		}

// 		claims, err := utils.ValidateToken(tokenParts[1])
// 		if err != nil {
// 			return c.Status(401).JSON(fiber.Map{
// 				"error": "Token tidak valid atau expired",
// 			})
// 		}

// 		// Simpan data user ke context
// 		c.Locals("user_id", claims.UserID.Hex())
// 		c.Locals("username", claims.Username)
// 		c.Locals("role", claims.Role)

// 		// Penting untuk alumni!
// 		if claims.AlumniID != "" {
// 			c.Locals("alumni_id", claims.AlumniID)
// 		}

// 		return c.Next()
// 	}
// }

// Middleware untuk memerlukan role admin
// func AdminOnly() fiber.Handler {
// 	return func(c *fiber.Ctx) error {
// 		roleValue := c.Locals("role")
// 		role, ok := roleValue.(string)
// 		if !ok || role != "admin" {
// 			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
// 				"error": "Akses ditolak, hanya admin yang diizinkan",
// 			})
// 		}
// 		return c.Next()
// 	}
// }

