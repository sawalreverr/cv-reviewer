package errors

import (
	"errors"
	"fmt"
)

var (
	// common error
	ErrNotFound = errors.New("resource not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrInternal = errors.New("internal server error")

	// document error
	ErrInvalidType = errors.New("invalid file type, only PDF allowed")
	ErrFileUpload = errors.New("failed to uplaod file")

	// job error
	ErrJobNotFound = errors.New("evaluation job not found")
	ErrJobFailed = errors.New("evaluation job failed")
	ErrJobTimeout = errors.New("evaluation job timeout")
)

type AppError struct {
	Code string
	Message string
	Err error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}

	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewAppError(code, message string, err error) *AppError {
	return &AppError{
		Code: code,
		Message: message,
		Err: err,
	}
}