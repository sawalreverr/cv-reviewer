package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type DocumentType string

const (
	CV DocumentType = "cv"
	ProjectReport DocumentType = "project_report"
	JobDescription DocumentType = "job_description"
	CaseStudyBrief DocumentType = "case_study_brief"
	CVRubric DocumentType = "cv_rubric"
	ProjectRubric DocumentType = "project_rubric"
)

// entity
type Document struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Type DocumentType `gorm:"type:text;not null;index" json:"type"`
	Filename string `gorm:"type:text;not null" json:"filename"`
	FilePath string `gorm:"type:text;not null" json:"file_path"`
	FileSize int64 `gorm:"not null" json:"file_size"`
	MimeType string `gorm:"type:text;not null" json:"mime_type"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"-"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"-"`
}

func NewDocument(docType DocumentType, filename, filepath string, filesize int64, mimetype string) *Document {
	return &Document{
		ID: uuid.New(),
		Type: docType,
		Filename: filename,
		FilePath: filepath,
		FileSize: filesize,
		MimeType: mimetype,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// contract
type DocumentRepository interface {
	Create(ctx context.Context, doc *Document) error
	FindByID(ctx context.Context, id uuid.UUID) (*Document, error)
	FindByType(ctx context.Context, docType DocumentType) ([]*Document, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

func (Document) TableName() string {
	return "documents"
}