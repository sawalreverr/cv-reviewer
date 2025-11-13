package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sawalreverr/cv-reviewer/internal/domain"
	"github.com/sawalreverr/cv-reviewer/pkg/errors"
	"gorm.io/gorm"
)

// evaluation job
type evaluationJobRepository struct {
	db *gorm.DB
} 

func NewEvaluationJobRepository(db *gorm.DB) domain.EvaluationJobRepository {
	return &evaluationJobRepository{db}
}

func (r *evaluationJobRepository) Create(ctx context.Context, job *domain.EvaluationJob) error {
	if err := r.db.WithContext(ctx).Create(job).Error; err != nil {
		return fmt.Errorf("failed to create evaluation job: %w", err)
	}

	return nil
}

func (r *evaluationJobRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.EvaluationJob, error) {
	var job domain.EvaluationJob
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&job).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrJobNotFound
		}

		return nil, fmt.Errorf("failed to find evaluation job: %w", err)
	}	

	return &job, nil
}

func (r *evaluationJobRepository) Update(ctx context.Context, job *domain.EvaluationJob) error {
	if err := r.db.WithContext(ctx).Save(&job).Error; err != nil {
		return fmt.Errorf("failed to update evaluation job: %w", err)
	}

	return nil
}

func (r *evaluationJobRepository) FindPendingJobs(ctx context.Context, limit int) ([]*domain.EvaluationJob, error) {
	var jobs []*domain.EvaluationJob
	if err := r.db.WithContext(ctx).Where("status = ?", domain.StatusQueued).Order("created_at ASC").Limit(limit).Find(&jobs).Error; err != nil {
		return nil, fmt.Errorf("failed to find pending jobs: %w", err)
	}

	return jobs, nil
}

// evaluation result
type evaluationResultRepository struct {
	db *gorm.DB
}

func NewEvaluationResultRepository(db *gorm.DB) domain.EvaluationResultRepository {
	return &evaluationResultRepository{db}
}

func (r *evaluationResultRepository) Create(ctx context.Context, result *domain.EvaluationResult) error {
	if err := r.db.WithContext(ctx).Create(result).Error; err != nil {
		return fmt.Errorf("failed to create evaluation result: %w", err)
	}

	return nil
}

func (r *evaluationResultRepository) FindByJobID(ctx context.Context, jobID uuid.UUID) (*domain.EvaluationResult, error) {
	var result domain.EvaluationResult
	if err := r.db.WithContext(ctx).Where("job_id = ?", jobID).First(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}

		return nil, fmt.Errorf("failed to find evaluation result: %w", err)
	}

	return &result, nil
} 