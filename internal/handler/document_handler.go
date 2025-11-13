package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sawalreverr/cv-reviewer/internal/domain"
	"github.com/sawalreverr/cv-reviewer/internal/usecase"
	"github.com/sawalreverr/cv-reviewer/pkg/errors"
	"github.com/sawalreverr/cv-reviewer/pkg/response"
)

type DocumentHandler struct {
	usecase usecase.DocumentUsecase
}

func NewDocumentHandler(uc usecase.DocumentUsecase) *DocumentHandler {
	return &DocumentHandler{uc}
}

// type DocumentInfo struct {
// 	ID uuid.UUID `json:"id"`
// 	Filename string `json:"filename"`
// 	FileSize int64 `json:"file_size"`
// 	Type string `json:"type"`
// }

type UploadResponse struct {
	CVID *uuid.UUID `json:"cv_id,omitempty"`
	ReportID *uuid.UUID `json:"report_id,omitempty"`
	// CVDocument *DocumentInfo `json:"cv_document,omitempty"`
	// ReportDocument *DocumentInfo `json:"report_document,omitempty"`
}

func (h *DocumentHandler) Upload(c echo.Context) error {
	ctx := c.Request().Context()

	form, err := c.MultipartForm()
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid form data", err)
	}

	cvFile := form.File["cv"]
	if len(cvFile) == 0 {
		return response.Error(c, http.StatusBadRequest, "cv file required", nil)
	}

	reportFile := form.File["report"]
	if len(reportFile) == 0 {
		return response.Error(c, http.StatusBadRequest, "report file required", nil)
	}

	cvDoc, err := h.usecase.UploadDocument(ctx, cvFile[0], domain.CV)
	if err != nil {
		return h.handleUploadError(c, err)
	}

	reportDoc, err := h.usecase.UploadDocument(ctx, reportFile[0], domain.ProjectReport)
	if err != nil {
		_ = h.usecase.DeleteDocument(ctx, cvDoc.ID)
		return h.handleUploadError(c, err)
	}

	resp := UploadResponse{
		CVID: &cvDoc.ID,
		ReportID: &reportDoc.ID,
		// CVDocument: &DocumentInfo{
		// 	ID:       cvDoc.ID,
		// 	Filename: cvDoc.Filename,
		// 	FileSize: cvDoc.FileSize,
		// 	Type:     string(cvDoc.Type),
		// },
		// ReportDocument: &DocumentInfo{
		// 	ID:       reportDoc.ID,
		// 	Filename: reportDoc.Filename,
		// 	FileSize: reportDoc.FileSize,
		// 	Type:     string(reportDoc.Type),
		// },
	}

	return response.Success(c, http.StatusCreated, "documents uploaded successfully", resp)
}

func (h *DocumentHandler) GetDocument(c echo.Context) error {
	ctx := c.Request().Context()

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid document id", err)
	}

	doc, err := h.usecase.GetDocument(ctx, id)
	if err != nil {
		if err == errors.ErrNotFound {
			return response.Error(c, http.StatusNotFound, "document not found", err)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to get document", err)
	}

	return response.SuccessData(c, doc)
}


func (h *DocumentHandler) handleUploadError(c echo.Context, err error) error {
	switch err {
	case errors.ErrInvalidType:
		return response.Error(c, http.StatusBadRequest, "invalid file type, only pdf allowed", err)
	default:
		return response.Error(c, http.StatusInternalServerError, "failed to upload document", err)
	}
}