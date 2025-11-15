package config

import (
	"fmt"
	"log"

	"github.com/sawalreverr/cv-reviewer/internal/domain"
	"gorm.io/gorm"
)

func RunMigration(db *gorm.DB) error {
	log.Println("running database migration...")

	// enable pgvector extension
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error; err != nil {
		return fmt.Errorf("failed to enable pgvector extension: %w", err)
	}

	entities := []interface{}{
		&domain.Document{},
		&domain.EvaluationJob{},
		&domain.EvaluationResult{},
		&domain.VectorDocument{},
	}

	// auto migrate tables
	if err := db.AutoMigrate(entities...); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// create vector index (hnsw/ivfflat)
	if err := createVectorIndex(db); err != nil {
		log.Printf("failed to create vector index: %v", err)
	}

	log.Println("database migrations completed!")
	return nil
}

func createVectorIndex(db *gorm.DB) error {
	// check if index already exists
	var indexExists bool
	checkQuery := `
	SELECT EXISTS (
		SELECT 1 
		FROM pg_indexes 
		WHERE tablename = 'vector_documents' 
		AND indexname = 'idx_vector_documents_embedding'
	)
	`

	if err := db.Raw(checkQuery).Scan(&indexExists).Error; err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}

	if indexExists {
		log.Println("vector index already exists")
		return nil
	}

	log.Println("creating vector index...")
	
	// create ivfflat index for faster similarity search
	createIndexQuery := `
		CREATE INDEX IF NOT EXISTS idx_vector_documents_embedding 
		ON vector_documents 
		USING ivfflat (embedding vector_cosine_ops)
		WITH (lists = 100)
		`
	if err := db.Exec(createIndexQuery); err != nil {
		log.Println("IVFFlat index creation failed (might need more data), trying basic index...")

		// fallback: create basic vector index without IVFFlat
		basicIndexQuery := `
			CREATE INDEX IF NOT EXISTS idx_vector_documents_embedding 
			ON vector_documents 
			USING ivfflat (embedding vector_cosine_ops)
		`
		if err := db.Exec(basicIndexQuery).Error; err != nil {
			return fmt.Errorf("failed to create vector index: %w", err)
		}
	}

	log.Println("vector index created!")
	return nil
}

func DropAllTables(db *gorm.DB) error {
	log.Println("dropping all tables...")

	entities := []interface{}{
		&domain.VectorDocument{},
		&domain.EvaluationResult{},
		&domain.EvaluationJob{},
		&domain.Document{},
	}

	if err := db.Migrator().DropTable(entities...); err != nil {
		return fmt.Errorf("failed to drop tables: %w", err)
	}

	log.Println("all tables dropped!")
	return nil
}