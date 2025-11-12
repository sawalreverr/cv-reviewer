package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/sawalreverr/cv-reviewer/pkg/response"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Check(c echo.Context) error {
	return response.SuccessData(c, map[string]string{
		"status": "ok",
		"message": "api is running",
	})
}