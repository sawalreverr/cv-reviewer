package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/sawalreverr/cv-reviewer/internal/domain"
	"github.com/sawalreverr/cv-reviewer/internal/service"
	"github.com/sawalreverr/cv-reviewer/pkg/errors"
)

type EvaluationUsecase interface {
	CreateEvaluationJob(ctx context.Context, jobTitle string, cvID, reportID uuid.UUID) (*domain.EvaluationJob, error)
	GetEvaluationJob(ctx context.Context, jobID uuid.UUID) (*domain.EvaluationJob, *domain.EvaluationResult, error)
	Process(ctx context.Context, job service.Job) error
}

type evaluationUsecase struct {
	jobRepo domain.EvaluationJobRepository
	resultRepo domain.EvaluationResultRepository
	documentRepo domain.DocumentRepository
	vectorUsecase VectorUsecase
	pdfService service.PDFService
	llmService service.LLMService
	jobTimeout time.Duration
}

func NewEvaluationUsecase(
	jobRepo domain.EvaluationJobRepository,
	resultRepo domain.EvaluationResultRepository,
	documentRepo domain.DocumentRepository,
	vectorUsecase VectorUsecase,
	pdfService service.PDFService,
	llmService service.LLMService,
	jobTimeout int,
) EvaluationUsecase {
	return &evaluationUsecase{
		jobRepo: jobRepo,
		resultRepo: resultRepo,
		documentRepo: documentRepo,
		vectorUsecase: vectorUsecase,
		pdfService: pdfService,
		llmService: llmService,
		jobTimeout: time.Duration(jobTimeout) * time.Second,
	}
}

func (uc *evaluationUsecase) CreateEvaluationJob(ctx context.Context, jobTitle string, cvID, reportID uuid.UUID) (*domain.EvaluationJob, error) {
	// validate document exist
	if _, err := uc.documentRepo.FindByID(ctx, cvID); err != nil {
		return nil, fmt.Errorf("cv document not found: %w", err)
	}

	if _, err := uc.documentRepo.FindByID(ctx, reportID); err != nil {
		return nil, fmt.Errorf("project report document not found: %w", err)
	}

	// create job
	job := domain.NewEvaluationJob(jobTitle, cvID, reportID)
	if err := uc.jobRepo.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create evaluation job: %w", err)	
	}

	return job, nil
}

func (uc *evaluationUsecase) GetEvaluationJob(ctx context.Context, jobID uuid.UUID) (*domain.EvaluationJob, *domain.EvaluationResult, error) {
	job, err := uc.jobRepo.FindByID(ctx, jobID)
	if err != nil {
		return nil, nil, err
	}

	// if completed, get result
	var result *domain.EvaluationResult
	if job.Status == domain.StatusCompleted {
		result, err = uc.resultRepo.FindByJobID(ctx, jobID)
		if err != nil && err != errors.ErrJobNotFound {
			return job, nil, err
		}
	}

	return job, result, nil
}

func (uc *evaluationUsecase) Process(ctx context.Context, job service.Job) error {
	// create timeout context
	timeoutCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// get job from db
	evalJob, err := uc.jobRepo.FindByID(ctx, job.ID)
	if err != nil {
		return fmt.Errorf("failed to find job: %w", err)
	}

	// mark as processing
	evalJob.MarkProcessing()
	if err := uc.jobRepo.Update(timeoutCtx, evalJob); err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	// process evaluation
	if err := uc.processEvaluation(timeoutCtx, evalJob); err != nil {
		// mark as failed, bcz err
		evalJob.MarkFailed(err.Error())
		uc.jobRepo.Update(context.Background(), evalJob)
		return err
	}

	// mark as completed
	evalJob.MarkCompleted()
	if err := uc.jobRepo.Update(context.Background(), evalJob); err != nil {
		return fmt.Errorf("failed to mark job as completed: %w", err)
	}

	return nil
}

func (uc *evaluationUsecase) processEvaluation(ctx context.Context, job *domain.EvaluationJob) error {
	log.Printf("[%s] -- processing evaluation", job.ID)

	// get cv document
	cvDoc, err := uc.documentRepo.FindByID(ctx, job.CVID)
	if err != nil {
		return fmt.Errorf("failed to get cv document: %w", err)
	}

	// get pr document
	reportDoc, err := uc.documentRepo.FindByID(ctx, job.ProjectReportID)
	if err != nil {
		return fmt.Errorf("failed to get project report document: %w", err)
	}

	// extract text from cv
	cvText, err := uc.pdfService.ExtractText(cvDoc.FilePath)
	if err != nil {
		return fmt.Errorf("failed to extract cv text: %s", err)
	}

	// extract text from pr
	prText, err := uc.pdfService.ExtractText(reportDoc.FilePath)
	if err != nil {
		return fmt.Errorf("failed to extract project text: %w", err)
	}

	// retrieve relevant job description context
	jdDocs, err := uc.vectorUsecase.SearchSimilar(ctx, job.JobTitle+" "+cvText[:min(500, len(cvText))], domain.JobDescription, 5)
	if err != nil {
		return fmt.Errorf("failed to search job description: %w", err)
	}
	jdContext := extractContent(jdDocs)

	// retrieve relevant cv scoring rubric
	cvRubricDocs, err := uc.vectorUsecase.SearchSimilar(ctx, "CV evaluation scoring criteria", domain.CVRubric, 3)
	if err != nil {
		return fmt.Errorf("failed to search cv rubric: %w", err)
	}
	cvRubricContext := extractContent(cvRubricDocs)

	// evaluate cv
	cvEval, err := uc.llmService.EvaluateCV(ctx, cvText, jdContext, cvRubricContext)
	if err != nil {
		return fmt.Errorf("failed to evaluate cv: %w", err)
	}

	// retrieve case study brief context 
	csDocs, err := uc.vectorUsecase.SearchSimilar(ctx, prText[:min(500, len(prText))], domain.CaseStudyBrief, 5)
	if err != nil {
		return fmt.Errorf("failed to search case study brief: %w", err)
	}
	csContext := extractContent(csDocs)

	// retrieve project scoring rubric
	projectRubricDocs, err := uc.vectorUsecase.SearchSimilar(ctx, "Project evaluation scoring criteria", domain.ProjectRubric, 3)
	if err != nil {
		return fmt.Errorf("failed to search project rubric: %w", err)
	}
	projectRubricContext := extractContent(projectRubricDocs)

	// evaluate project
	projectEval, err := uc.llmService.EvaluateProject(ctx, prText, csContext, projectRubricContext)
	if err != nil {
		return fmt.Errorf("failed to evaluate project: %w", err)
	}

	// final summary
	summary, err := uc.llmService.FinalSummary(ctx, cvEval, projectEval)
	if err != nil {
		return fmt.Errorf("failed to generate summary: %w", err)
	}

	// save result
	result := domain.NewEvaluationResult(job.ID, cvEval.CVMatchRate, cvEval.CVFeedback, projectEval.ProjectScore, projectEval.ProjectFeedback, summary)
	if err := uc.resultRepo.Create(ctx, result); err != nil {
		return fmt.Errorf("failed to save evaluation result: %w", err)
	}

	log.Printf("[%s] -- evaluation completed", job.ID)
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func extractContent(docs []*domain.VectorDocument) []string {
	contents := make([]string, len(docs))
	for i, doc := range docs {
		contents[i] = doc.Content
	}
	return contents
}