package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sawalreverr/cv-reviewer/internal/domain"
	"github.com/sawalreverr/cv-reviewer/internal/service"
	"github.com/sawalreverr/cv-reviewer/internal/usecase"
	"github.com/sawalreverr/cv-reviewer/pkg/errors"
	"github.com/sawalreverr/cv-reviewer/pkg/response"
)

type EvaluationHandler struct {
	usecase usecase.EvaluationUsecase
	jobQueue service.JobQueue
}

func NewEvaluationHandler(uc usecase.EvaluationUsecase, jobQueue service.JobQueue) *EvaluationHandler {
	return &EvaluationHandler{uc, jobQueue}
}

type EvaluateRequest struct {
	JobTitle string `json:"job_title" validate:"required"`
	CVID uuid.UUID `json:"cv_id" validate:"required"`
	ProjectReportID uuid.UUID `json:"project_report_id" validate:"required"`
}

type EvaluateResponse struct {
	ID uuid.UUID `json:"id"`
	Status string `json:"status"`
}

func (h *EvaluationHandler) Evaluate(c echo.Context) error {
	ctx := c.Request().Context()

	var req EvaluateRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", err)
	}

	// validate requeired field
	if req.JobTitle == "" {
		return response.Error(c, http.StatusBadRequest, "job_title is required", nil)
	}
	if req.CVID == uuid.Nil {
		return response.Error(c, http.StatusBadRequest, "cv_id is required", nil)
	}
	if req.ProjectReportID == uuid.Nil {
		return response.Error(c, http.StatusBadRequest, "project_report_id is required", nil)
	}

	// create evaluation job
	job, err := h.usecase.CreateEvaluationJob(ctx, req.JobTitle, req.CVID, req.ProjectReportID)
	if err != nil {
		if err == errors.ErrNotFound {
			return response.Error(c, http.StatusNotFound, "cv or project report document not found", err)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to create evaluation job", err)
	}
	
	// enqueue job for async processing
	if err := h.jobQueue.Enqueue(service.Job{
		ID: job.ID,
		JobTitle: job.JobTitle,
		CVID: job.CVID,
		ProjectID: job.ProjectReportID,
	}); err != nil {
		return response.Error(c, http.StatusServiceUnavailable, "job queue is full, please try again later", err)
	}

	resp := EvaluateResponse{
		ID: job.ID,
		Status: string(job.Status),
	}

	return response.Success(c, http.StatusCreated, "evaluation job created", resp)
}

type ResultResponse struct {
	ID uuid.UUID `json:"id"`
	Status string `json:"status"`
	Result *ResultData `json:"result,omitempty"`
}

type ResultData struct {
	CVMatchRate float64 `json:"cv_match_rate"`
	CVFeedback string `json:"cv_feedback"`
	ProjectScore float64 `json:"project_score"`
	ProjectFeedback string `json:"project_feedback"`
	OverallSummary string `json:"overall_summary"`
}

func (h *EvaluationHandler) GetResult(c echo.Context) error {
	ctx := c.Request().Context()

	// parse job id
	idStr := c.Param("id")
	jobID, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid job id", err)
	}
	
	// get job and result
	job, result, err := h.usecase.GetEvaluationJob(ctx ,jobID)
	if err != nil {
		if err == errors.ErrJobNotFound {
			return response.Error(c, http.StatusNotFound, "evaluation job not found", err)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to get evaluation job", err)
		
	}

	resp := ResultResponse{
		ID: jobID,
		Status: string(job.Status),
	}

	// if completed
	if job.Status == domain.StatusCompleted && result != nil {
		resp.Result = &ResultData{
			CVMatchRate: result.CVMatchRate,
			CVFeedback: result.CVFeedback,
			ProjectScore: result.ProjectScore,
			ProjectFeedback: result.ProjectFeedback,
			OverallSummary: result.OverallSummary,
		}
	}

	return response.SuccessData(c, resp)

}