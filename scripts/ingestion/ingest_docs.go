package main

import (
	"context"
	"log"
	"os"

	"github.com/sawalreverr/cv-reviewer/config"
	"github.com/sawalreverr/cv-reviewer/internal/domain"
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

	// connect db
	db, err := config.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// run migrations
	if err := config.RunMigration(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// init services
	pdfService := service.NewPDFService()
	chunkingService := service.NewChunkingService()
	embeddingService, err := service.NewEmbeddingService(&cfg.Gemini)
	if err != nil {
		log.Fatalf("failed to create embedding service: %v", err)
	}

	// init repo
	vectorRepo := repository.NewVectorRepository(db)
	vectorUsecase := usecase.NewVectorUsecase(vectorRepo, pdfService, chunkingService, embeddingService)

	ctx := context.Background()
	systemDocs := []struct{
		Path string
		Doctype domain.DocumentType
		Metadata map[string]interface{}
	} {
		{
			Path: "./docs/job_description.pdf",
			Doctype: domain.JobDescription,
			Metadata: map[string]interface{}{
				"source": "job_description",
				"description": "Backend Product Engineer Job Description",
				"version": "1.0",
			},
		},
		{
			Path: "./docs/case_study_brief.pdf",
			Doctype: domain.CaseStudyBrief,
			Metadata: map[string]interface{}{
				"source": "case_study",
				"description": "Case Study Brief for Backend Developer Assesment",
				"version": "1.0",
			},
		},
		{
			Path: "./docs/cv_scoring_rubric.pdf",
			Doctype: domain.CVRubric,
			Metadata: map[string]interface{}{
				"source": "cv_rubric",
				"description": "CV Evaluation Scoring Rubric",
				"version": "1.0",
			},
		},
		{
			Path: "./docs/project_scoring_rubric.pdf",
			Doctype: domain.ProjectRubric,
			Metadata: map[string]interface{}{
				"source": "project_rubric",
				"description": "Project Deliverable Evaluation Scoring Rubric",
				"version": "1.0",
			},
		},
	}

	successCount := 0
	for _, doc := range systemDocs{
		log.Printf("ingesting: %s\ntype: %s", doc.Path, doc.Doctype)

		// check if file exist
		if _, err := os.Stat(doc.Path); os.IsNotExist(err) {
			log.Printf(" file not found: %s", doc.Path)
			continue
		}

		// delete existing documents
		log.Printf("cleaning existing document by type: %s", doc.Doctype)
		if err := vectorUsecase.DeleteDocumentsByType(ctx, doc.Doctype); err != nil {
			log.Printf("failed to delete existing document: %v", err)
		}

		// ingest document
		if err := vectorUsecase.IngestDocument(ctx, doc.Path, doc.Doctype, doc.Metadata); err != nil {
			log.Printf("failed to ingest %s: %v", doc.Path, err)
		}

		count, _ := vectorUsecase.GetDocumentCount(ctx, doc.Doctype)
		log.Printf("successfully ingested: %s (%d chunks stored)\n", doc.Path, count)
		successCount++
	}

	log.Printf("\ningestion completed: %d/%d documents succesful", successCount, len(systemDocs))
}	