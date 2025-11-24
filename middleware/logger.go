package middleware

import (
    "fmt"
    "github.com/gofiber/fiber/v2"
)

func LoggerMiddleware(c *fiber.Ctx) error {
    fmt.Printf("Request: %s %s\n", c.Method(), c.OriginalURL())
    err := c.Next()
    fmt.Printf("Response: %d\n", c.Response().StatusCode())
    return err
}