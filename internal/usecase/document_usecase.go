package usecase

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/sawalreverr/cv-reviewer/config"
	"github.com/sawalreverr/cv-reviewer/internal/domain"
	"github.com/sawalreverr/cv-reviewer/pkg/errors"
)

type DocumentUsecase interface {
	UploadDocument(ctx context.Context, file *multipart.FileHeader, docType domain.DocumentType) (*domain.Document, error)
	GetDocument(ctx context.Context, id uuid.UUID) (*domain.Document, error)
	DeleteDocument(ctx context.Context, id uuid.UUID) error
}

type documentUsecase struct {
	repo domain.DocumentRepository
	cfg *config.StorageConfig
}

func NewDocumentUsecase(repo domain.DocumentRepository, cfg *config.StorageConfig) DocumentUsecase {
	return &documentUsecase{repo, cfg}
}

func (uc *documentUsecase) UploadDocument(ctx context.Context, file *multipart.FileHeader, docType domain.DocumentType) (*domain.Document, error) {
	// try open uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, errors.NewAppError("FILE_OPEN_ERROR", "failed to open uploaded file", err)
	}
	defer src.Close()
	
	// validate file type
	if !uc.isPdf(file) {
		return nil, errors.ErrInvalidType
	}

	// reset pointer file, agar file yang disimpan tidak corrupt 
	if _, err := src.Seek(0, 0); err != nil {
        return nil, errors.NewAppError("FILE_SEEK_ERROR", "failed to reset file pointer", err)
    }

	ext := filepath.Ext(file.Filename)
	newFilename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	uploadPath := filepath.Join(uc.cfg.UploadDir, string(docType))
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		return nil, errors.NewAppError("DIR_CREATE_ERROR", "failed to create upload directory", err)
	}

	// create new file with generate id
	newFilePath := filepath.Join(uploadPath, newFilename)
	newFile, err := os.Create(newFilePath)
	if err != nil {
		return nil, errors.NewAppError("FILE_CREATE_ERROR", "failed to create file", err)
	}
	defer newFile.Close()

	// copy uploaded file content to newfile	
	if _, err := io.Copy(newFile, src); err != nil {
		return nil, errors.NewAppError("FILE_COPY_ERROR", "failed to copy file content", err)
	}

	doc := domain.NewDocument(docType, file.Filename, newFilePath, file.Size, file.Header.Get("Content-Type"))

	// save to db
	if err := uc.repo.Create(ctx, doc); err != nil {
		os.Remove(newFilePath)
		return nil, err
	}

	return doc, err
}

func (uc *documentUsecase) GetDocument(ctx context.Context, id uuid.UUID) (*domain.Document, error) {
	return uc.repo.FindByID(ctx, id)
}

func (uc *documentUsecase) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	// file exist ?
	doc, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// db delete
	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}

	// delete file from storage
	if err := os.Remove(doc.FilePath); err != nil && !os.IsNotExist(err) {
		return errors.NewAppError("FILE_DELETE_ERROR", "failed to delete file", err)
	}

	return nil
}

// validate pdf file
func (uc *documentUsecase) isPdf(file *multipart.FileHeader) bool {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".pdf" {
		return false
	}

	contentType := file.Header.Get("Content-Type")
	return contentType == "application/pdf"
}