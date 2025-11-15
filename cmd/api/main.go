package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sawalreverr/cv-reviewer/config"
	"github.com/sawalreverr/cv-reviewer/internal/handler"
	"github.com/sawalreverr/cv-reviewer/internal/repository"
	"github.com/sawalreverr/cv-reviewer/internal/usecase"
)

func main() {
	// load config
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }

	// init db
	db, err := config.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	
	// run migrations
	if err := config.RunMigration(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// init repositories
	documentRepo := repository.NewDocumentRepository(db)

	// init usecases
	documentUsecase := usecase.NewDocumentUsecase(documentRepo, &cfg.Storage)

	// init handlers
	healthHandler := handler.NewHealthHandler()
	documentHandler := handler.NewDocumentHandler(documentUsecase)

	// init echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())

	// routes
	e.GET("/health", healthHandler.Check)
	e.POST("/upload", documentHandler.Upload)
	
	log.Printf("server starting on port %s", cfg.Server.Port)
	if err := e.Start(":" + cfg.Server.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}