package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sawalreverr/cv-reviewer/config"
	"github.com/sawalreverr/cv-reviewer/internal/handler"
)

func main() {
	// load config
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }

	// init db
	_, err = config.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// init echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())

	// routes
	healthHandler := handler.NewHealthHandler()
	e.GET("/health", healthHandler.Check)

	log.Printf("server starting on port %s", cfg.Server.Port)
	if err := e.Start(":" + cfg.Server.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}