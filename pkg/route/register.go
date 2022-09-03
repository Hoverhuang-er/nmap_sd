package route

import "github.com/gofiber/fiber/v2"

func RegistHTTP() error {
	fb := fiber.New()
	fb.Get("/scan", GetScanResult)
	return fb.Listen(":8080")
}