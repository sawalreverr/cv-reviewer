package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Response struct {
	Success bool `json:"success"`
	Message string `json:"message,omitempty"` // omitempty itu klo nilainya kosong maka fieldnya hilang
	Data interface{} `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

func Success(c echo.Context, statusCode int, message string, data interface{}) error {
	return c.JSON(statusCode, Response{
		Success: true,
		Message: message,
		Data: data,
	})
}

func Error(c echo.Context, statusCode int, message string, err error) error {
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}

	return c.JSON(statusCode, Response{
		Success: false,
		Message: message,
		Error: errMsg,
	}) 
}

func SuccessData(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, Response{
		Success: true,
		Data: data,
	})
}