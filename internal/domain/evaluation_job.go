package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type JobStatus string

const (
	StatusQueued JobStatus = "queued"
	StatusProcessing JobStatus = "processing"
	StatusCompleted JobStatus = "completed"
	StatusFailed JobStatus = "failed"
)

// entity
type EvaluationJob struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	JobTitle string `gorm:"type:text;not null" json:"job_title"`
	CVID uuid.UUID `gorm:"type:uuid;not null;index" json:"cv_id"`
	ProjectReportID uuid.UUID `gorm:"type:uuid;not null" json:"project_report_id"`
	Status JobStatus `gorm:"type:text;not null;index" json:"status"`
	ErrorMessage *string `gorm:"type:text;default:null" json:"error_message,omitempty"` // optional, bisa nil
	StartedAt *time.Time `gorm:"type:timestamptz;default:null" json:"started_at,omitempty"` // optional, bisa nil
	CompletedAt *time.Time `gorm:"type:timestamptz;default:null" json:"completed_at,omitempty"` // optional, bisa nil
	CreatedAt time.Time `gorm:"autoCreateTime" json:"-"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"-"`
}

func NewEvaluationJob(jobTitle string, cvID, projectReportID uuid.UUID) *EvaluationJob {
	return &EvaluationJob{
		ID: uuid.New(),
		JobTitle: jobTitle,
		CVID: cvID,
		ProjectReportID: projectReportID,
		Status: StatusQueued,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (ej *EvaluationJob) MarkProcessing() {
	now := time.Now()
	ej.Status = StatusProcessing
	ej.StartedAt = &now
	ej.UpdatedAt = now
}

func (ej *EvaluationJob) MarkCompleted() {
	now := time.Now()
	ej.Status = StatusCompleted
	ej.CompletedAt = &now
	ej.UpdatedAt = now
}

func (ej *EvaluationJob) MarkFailed(msg string) {
	now := time.Now()
	ej.Status = StatusFailed
	ej.ErrorMessage = &msg
	ej.CompletedAt = &now
	ej.UpdatedAt = now
}

// contract
type EvaluationJobRepository interface {
	Create(ctx context.Context, job *EvaluationJob) error
	FindByID(ctx context.Context, id uuid.UUID) (*EvaluationJob, error)
	Update(ctx context.Context, job *EvaluationJob) error
	FindPendingJobs(ctx context.Context, limit int) ([]*EvaluationJob, error)
}

func (EvaluationJob) TableName() string {
	return "evaluation_jobs"
}