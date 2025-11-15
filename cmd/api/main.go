package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sawalreverr/cv-reviewer/config"
	"github.com/sawalreverr/cv-reviewer/internal/handler"
	"github.com/sawalreverr/cv-reviewer/internal/repository"
	"github.com/sawalreverr/cv-reviewer/internal/service"
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
	evaluationJobRepo := repository.NewEvaluationJobRepository(db)
	evaluationResultRepo := repository.NewEvaluationResultRepository(db)
	vectorRepo := repository.NewVectorRepository(db)

	// init services
	pdfService := service.NewPDFService()
	chunkingService := service.NewChunkingService()

	embeddingService, err := service.NewEmbeddingService(&cfg.Gemini)
	if err != nil {
		log.Fatalf("failed to create embedding service: %v", err)
	}
	
	llmService, err := service.NewLLMService(&cfg.Gemini)
	if err != nil {
		log.Fatalf("failed to create llm service: %v", err)
	}

	// init usecases
	documentUsecase := usecase.NewDocumentUsecase(documentRepo, &cfg.Storage)
	vectorUsecase := usecase.NewVectorUsecase(vectorRepo, pdfService, chunkingService, embeddingService)
	evaluationUsecase := usecase.NewEvaluationUsecase(evaluationJobRepo, evaluationResultRepo, documentRepo, vectorUsecase, pdfService, llmService, cfg.Queue.JobTimeout)

	// init job queue
	jobQueue := service.NewJobQueue(&cfg.Queue, evaluationUsecase)

	// start workers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	jobQueue.Start(ctx)

	// init handlers
	healthHandler := handler.NewHealthHandler()
	documentHandler := handler.NewDocumentHandler(documentUsecase)
	evaluationHandler := handler.NewEvaluationHandler(evaluationUsecase, jobQueue)

	// init echo
	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "[${method} | ${status}] ~ ${uri} ~ ${remote_ip} ~ ${latency_human}\n",
	}))
	e.Use(middleware.CORS())

	// routes
	e.GET("/health", healthHandler.Check)
	e.POST("/upload", documentHandler.Upload)
	e.POST("/evaluate", evaluationHandler.Evaluate)
	e.GET("/result/:id", evaluationHandler.GetResult)

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func(){
		<-quit
		log.Println("shutting down server...")

		// stop job queue
		jobQueue.Stop()

		// shutdown server with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := e.Shutdown(ctx); err != nil {
			log.Printf("server forced to shutdown: %v", err)
		}
	}()
	
	log.Printf("server starting on port %s", cfg.Server.Port)
	if err := e.Start(":" + cfg.Server.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}