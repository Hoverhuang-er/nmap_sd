package route

import (
	"os"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
)

func IndexPage(ctx *fiber.Ctx) error {
	return ctx.SendString("hello world")
}
func DisplayMgSDInfo(ctx *fiber.Ctx) error {
	f2, err := os.ReadFile("mg_sd_dip.json")
	if err != nil {
		log.Errorf("failed to read file: %v", err)
		return ctx.SendString("failed to read file")
	}
	return ctx.Send(f2)
}
