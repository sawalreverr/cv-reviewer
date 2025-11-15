package domain

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

type JSONB map[string]interface{}

// impl sql.Scanner
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = make(map[string]interface{})
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	result := make(map[string]interface{})
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}

	*j = result
	return nil
}

// impl driver.Valuer
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// entity
type VectorDocument struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	DocType DocumentType `gorm:"type:text;not null;index" json:"doc_type"`
	Content string `gorm:"type:text;not null" json:"content"`
	Embedding pgvector.Vector `gorm:"type:vector(768)" json:"-"` 
	Metadata  JSONB `gorm:"type:jsonb" json:"metadata"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"-"`
}

func NewVectorDocument(docType DocumentType, content string, embedding []float32, metadata map[string]interface{}) *VectorDocument {
	return &VectorDocument{
		ID: uuid.New(),
		DocType: docType,
		Content: content,
		Embedding: pgvector.NewVector(embedding),
		Metadata: JSONB(metadata),
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