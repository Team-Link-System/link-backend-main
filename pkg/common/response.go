package common

import (
	"fmt"
	"link/pkg/logger"
	"log"
	"net/http"
)

type Response struct {
	StatusCode int         `json:"statusCode"`
	Message    string      `json:"message"`
	Success    bool        `json:"success"`
	Payload    interface{} `json:"payload,omitempty"`
	Err        error       `json:"-"`
}

type AppError struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Success    bool   `json:"success"`
	Err        error  `json:"error,omitempty"`
}

func (r *Response) SuccessResponse() interface{} {
	return Response{
		StatusCode: http.StatusOK,
		Message:    r.Message,
		Success:    true,
		Payload:    r.Payload,
	}
}

func NewResponse(status int, message string, payload interface{}) Response {

	return Response{
		StatusCode: status,
		Message:    message,
		Success:    true,
		Payload:    payload,
	}
}

func (e *AppError) Error() string {
	if e.Err != nil {
		log.Printf("error: %v", e.Err) //에러 로그는 서버에만
		return fmt.Sprintf("%v", e.Message)
	}
	return e.Message
}

func NewError(status int, message string, err error) *AppError {
	appErr := &AppError{
		StatusCode: status,
		Success:    false,
		Message:    message,
	}

	logger.LogError(fmt.Sprintf("[%d] %s: %v", appErr.StatusCode, appErr.Message, appErr.Err))
	return appErr
}
