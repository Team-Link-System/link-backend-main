package common

import (
	"fmt"
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
	Err        error  `json:"-"`
}

// func Success(message string, data interface{}) Response {
// 	return Response{
// 		StatusCode: http.StatusOK,
// 		Message:    message,
// 		Success:    true,
// 		Payload:    data,
// 	}
// }

// func Created(message string, data interface{}) Response {
// 	return Response{
// 		StatusCode: http.StatusCreated,
// 		Message:    message,
// 		Success:    true,
// 		Payload:    data,
// 	}
// }

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

func NewError(status int, message string) *AppError {
	return &AppError{
		StatusCode: status,
		Success:    false,
		Message:    message,
	}
}
