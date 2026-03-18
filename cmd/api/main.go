package main

import (
	"org-api/config"
	"org-api/internal/app"
	"org-api/pkg/logger"
)

func main() {
	log := logger.New()
	cfg := config.Load()
 
	a, err := app.New(cfg, log)
	if err != nil {
		log.Fatalf("init app: %v", err)
	}
 
	if err := a.Run(); err != nil {
		log.Fatalf("run app: %v", err)
	}
}
 