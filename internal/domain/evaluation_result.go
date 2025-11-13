package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// entity
type EvaluationResult struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	JobID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"job_id"`
	CVMatchRate float64 `gorm:"not null" json:"cv_match_rate"`
	CVFeedback string `gorm:"type:text;not null" json:"cv_feedback"`
	ProjectScore float64 `gorm:"not null" json:"project_score"`
	ProjectFeedback string `gorm:"type:text;not null" json:"project_feedback"`
	OverallSummary string `gorm:"type:text;not null" json:"overall_summary"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"-"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"-"`
}

func NewEvaluationResult(jobID uuid.UUID, cvMatchRate float64, cvFeedback string, projectScore float64, projectFeedback string, overallSummary string) *EvaluationResult {
	return &EvaluationResult{
		ID: uuid.New(),
		JobID: jobID,
		CVMatchRate: cvMatchRate,
		CVFeedback: cvFeedback,
		ProjectScore: projectScore,
		ProjectFeedback: projectFeedback,
		OverallSummary: overallSummary,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// contract
type EvaluationResultRepository interface {
	Create(ctx context.Context, result *EvaluationResult) error
	FindByJobID(ctx context.Context, jobID uuid.UUID) (*EvaluationResult, error)
}

func (EvaluationResult) TableName() string {
	return "evaluation_results"
}