package repository

import (
	"context"
	"fmt"

	"github.com/pgvector/pgvector-go"
	"github.com/sawalreverr/cv-reviewer/internal/domain"
	"gorm.io/gorm"
)

type vectorRepository struct {
	db *gorm.DB
}

func NewVectorRepository(db *gorm.DB) domain.VectorRepository {
	return &vectorRepository{db}
}

func (r *vectorRepository) Create(ctx context.Context, doc *domain.VectorDocument) error {
	if err := r.db.WithContext(ctx).Create(doc).Error; err != nil {
		return fmt.Errorf("failed to create vector documents: %w", err)
	}

	return nil
}

func (r *vectorRepository) SearchSimilar(ctx context.Context, embedding []float32, docType domain.DocumentType, limit int) ([]*domain.VectorDocument, error) {
	var docs []*domain.VectorDocument

	pgEmbedding := pgvector.NewVector(embedding)

	// cosine similarity (skor similar 0-1), order by distance closest to farthest
	query := r.db.WithContext(ctx).Select("*, 1 - (embedding <=> ?) as similarity", pgEmbedding).Where("doc_type = ?", docType).Order(gorm.Expr("embedding <=> ?", pgEmbedding)).Limit(limit)

	if err := query.Find(&docs).Error; err != nil {
		return nil, fmt.Errorf("failed to search similar documents: %w", err)
	}

	return docs, nil
}

func (r *vectorRepository) DeleteByDocType(ctx context.Context, docType domain.DocumentType) error {
	if err := r.db.WithContext(ctx).Where("doc_type = ?", docType).Delete(&domain.VectorDocument{}).Error; err != nil {
		return fmt.Errorf("failed to delete vector documents: %w", err)
	}
	return nil
}

func (r *vectorRepository) Count(ctx context.Context, docType domain.DocumentType) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.VectorDocument{}).Where("doc_type = ?", docType).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count vector documents: %w", err)
	}

	return count, nil
}