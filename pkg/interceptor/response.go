package interceptor

import "net/http"

type Response struct {
	StatusCode int         `json:"statusCode"`
	Message    string      `json:"message"`
	Success    bool        `json:"success"`
	Payload    interface{} `json:"payload,omitempty"`
}

func Success(message string, data interface{}) Response {
	return Response{
		StatusCode: http.StatusOK,
		Message:    message,
		Success:    true,
		Payload:    data,
	}
}

func Created(message string, data interface{}) Response {
	return Response{
		StatusCode: http.StatusCreated,
		Message:    message,
		Success:    true,
		Payload:    data,
	}
}

func Error(status int, message string) Response {
	return Response{
		StatusCode: status,
		Success:    false,
		Message:    message,
	}
}
