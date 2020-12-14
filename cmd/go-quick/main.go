package main

import (
	"go-quick/app"
	"go-quick/config"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initiate a simple logger
	log := logrus.New()

	// Load Config from environment
	env, err := config.NewFromEnv()
	if err != nil {
		log.Fatalf("Unable to load config shutting down - %s", err)
	}

	// Run application
	err = app.Run(env)
	if err != nil && err != app.ErrShutdown {
		log.Fatalf("Service stopped - %s", err)
	}
	log.Infof("Service shutdown - %s", err)
}
