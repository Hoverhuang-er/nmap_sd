package main

import (
	"nmap_sd/pkg/route"
	"nmap_sd/pkg/sd"
	"time"

	"github.com/go-co-op/gocron"
)
func main()  {
	// Regist Query and Command from remote
	route.RegistHTTP()
	// Register cron job with go-corn
	tc := gocron.NewScheduler(time.Local)
	tc.Every(1).Day().At("00:00").Do(sd.ScanAll())
	tc.StartBlocking()
}
