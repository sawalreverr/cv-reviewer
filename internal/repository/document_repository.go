package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sawalreverr/cv-reviewer/internal/domain"
	"github.com/sawalreverr/cv-reviewer/pkg/errors"
	"gorm.io/gorm"
)

type documentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) domain.DocumentRepository {
	return &documentRepository{db}
}

func (r *documentRepository) Create(ctx context.Context, doc *domain.Document) error {
	if err := r.db.WithContext(ctx).Create(doc).Error; err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}

	return nil
}

func (r *documentRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Document, error) {
	var doc domain.Document
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&doc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}

		return nil, fmt.Errorf("failed to find document: %w", err)
	}

	return &doc, nil
}

func (r *documentRepository) FindByType(ctx context.Context, docType domain.DocumentType) ([]*domain.Document, error) {
	var docs []*domain.Document
	if err := r.db.WithContext(ctx).Where("type = ?", docType).Find(&docs).Error; err != nil {
		return nil, fmt.Errorf("failed to find documents by type: %w", err)
	}

	return docs, nil
}


func (r *documentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.Document{}).Error; err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}