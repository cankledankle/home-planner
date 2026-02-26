package middleware

import "github.com/gofiber/fiber/v2"

func AdminMiddleware(c *fiber.Ctx) error {
	userRole := c.Locals("userRole")

	if userRole == nil || userRole != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "FORBIDDEN",
				"message": "Admin access required",
			},
		})
	}

	return c.Next()
}
