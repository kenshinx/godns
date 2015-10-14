package main

import (
	"os"
	"os/signal"
	"runtime"
	"time"
)

var (
	logger *GoDNSLogger
)

func main() {

	initLogger()

	server := &Server{
		host:     settings.Server.Host,
		port:     settings.Server.Port,
		rTimeout: 5 * time.Second,
		wTimeout: 5 * time.Second,
	}

	server.Run()

	logger.Info("godns %s start", settings.Version)

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

forever:
	for {
		select {
		case <-sig:
			logger.Info("signal received, stopping")
			break forever
		}
	}

}

func initLogger() {
	logger = NewLogger()

	if settings.Log.Stdout {
		logger.SetLogger("console", nil)
	}

	if settings.Log.File != "" {
		config := map[string]interface{}{"file": settings.Log.File}
		logger.SetLogger("file", config)
	}

	logger.SetLevel(settings.Log.LogLevel())
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
