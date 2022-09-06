package main

import (
	"fmt"
	"nmap_sd/pkg/route"
	"nmap_sd/pkg/sd"
	"math/rand"
	"time"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/go-co-op/gocron"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
)

func main() {
	gc := gocron.NewScheduler(time.Local)
	gc.Every(1).Minute().Do(sd.ScanNetworkWNmap)
	gc.StartAsync()
	app := fiber.New(fiber.Config{
		JSONEncoder: jsoniter.Marshal,
		JSONDecoder: jsoniter.Unmarshal,
	})
	prometheus := fiberprometheus.New("mg_sd")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)
	app.Use(pprof.New())
	app.Use(recover.New())
	app.Use(cache.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "https://fjmg.online, https://fjmg.com",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed, // 1
	}))
	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))
	app.Get("/", route.IndexPage)
	app.Get("/dip", route.DisplayMgSDInfo)
	app.Get("/monitor", monitor.New(monitor.Config{Title: "PageInternalMonitor"}))
	rport := fmt.Sprintf(":%d", rand.Int63n(65535))
	if err := app.Listen(rport); err != nil {
		log.Errorf("failed to start fiber: %v", err)
		return
	}
}
