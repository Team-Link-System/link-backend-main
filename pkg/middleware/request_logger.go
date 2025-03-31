package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func RequestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// 바디를 다시 설정
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		w := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = w

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		userID, exists := c.Get("userId")

		// zap 로거 필드 구성
		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.String("ip", clientIP),
			zap.Duration("duration", duration),
			zap.Int64("duration_ms", duration.Milliseconds()),
		}

		if exists {
			fields = append(fields, zap.Any("user_id", userID))
		}

		// 상태 코드에 따른 로그 레벨 설정
		switch {
		case statusCode >= 500:
			logger.Error("API 요청 실패", fields...)
		case statusCode >= 400:
			logger.Warn("API 요청 오류", fields...)
		default:
			logger.Info("API 요청 성공", fields...)
		}
	}
}
