package middleware

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"link/pkg/logger"

	"github.com/gin-gonic/gin"
)

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 시작 시간 기록
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// 요청 바디 읽기
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// 바디를 다시 설정
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 응답 캡처를 위한 커스텀 writer
		w := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = w

		// 다음 핸들러 실행
		c.Next()

		// 요청 처리 완료 후
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		userID, _ := c.Get("userId") // 인증된 사용자 ID

		// 로그 메시지 구성
		logMessage := fmt.Sprintf("[%s] %s | Status: %d | IP: %s | Duration: %v",
			method,
			path,
			statusCode,
			clientIP,
			duration,
		)
		if userID != nil {
			logMessage += fmt.Sprintf(" | UserID: %v", userID)
		}

		// 상태 코드에 따른 로그 레벨 설정
		switch {
		case statusCode >= 500:
			logger.LogError(logMessage)
		case statusCode >= 400:
			logger.LogError(logMessage)
		default:
			logger.LogSuccess(logMessage) // Success 대신 Info 레벨 사용
		}
	}
}
