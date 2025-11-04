package logutil

import (
	"log"
	"os"
)

var (
	LogLevel = getLogLevel()
)

func getLogLevel() string {
	lvl := os.Getenv("LOG_LEVEL")
	if lvl == "" {
		return "debug" // default to debug for dev
	}
	return lvl
}

func Debugf(format string, v ...interface{}) {
	if LogLevel == "debug" {
		log.Printf("[DEBUG] "+format, v...)
	}
}

func Infof(format string, v ...interface{}) {
	if LogLevel == "debug" || LogLevel == "info" {
		log.Printf("[INFO] "+format, v...)
	}
}

func Warnf(format string, v ...interface{}) {
	if LogLevel == "debug" || LogLevel == "info" || LogLevel == "warn" {
		log.Printf("[WARN] "+format, v...)
	}
}

func Errorf(format string, v ...interface{}) {
	if LogLevel != "none" {
		log.Printf("[ERROR] "+format, v...)
	}
}
