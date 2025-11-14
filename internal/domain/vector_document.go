package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

// entity
type VectorDocument struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	DocType DocumentType `gorm:"type:text;not null;index" json:"doc_type"`
	Content string `gorm:"type:text;not null" json:"content"`
	Embedding pgvector.Vector `gorm:"type:vector(768)" json:"-"` 
	Metadata  map[string]interface{} `gorm:"type:jsonb" json:"metadata"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"-"`
}

func NewVectorDocument(docType DocumentType, content string, embedding []float32, metadata map[string]interface{}) *VectorDocument {
	return &VectorDocument{
		ID: uuid.New(),
		DocType: docType,
		Content: content,
		Embedding: pgvector.NewVector(embedding),
		Metadata: metadata,
		CreatedAt: time.Now(),
	}
}

// contract
type VectorRepository interface{
	Create(ctx context.Context, doc *VectorDocument) error
	SearchSimilar(ctx context.Context, embedding []float32, docType DocumentType, limit int) ([]*VectorDocument, error)
	DeleteByDocType(ctx context.Context, docType DocumentType) error
	Count(ctx context.Context, docType DocumentType) (int64, error)
}

func (VectorDocument) TableName() string {
	return "vector_documents"
}