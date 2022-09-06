package util

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func init() {
	// register logs to file
	logf, err := os.OpenFile("mg_sd.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(logf)
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(logf)
	log.SetLevel(log.DebugLevel)
}
