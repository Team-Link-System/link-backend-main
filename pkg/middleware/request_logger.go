package middleware

import (
	"bytes"
	"fmt"
	"time"

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
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		rw := &responseWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = rw

		// 요청 처리
		c.Next()

		// 응답 정보 로깅
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// 사용자 ID 가져오기 (있는 경우)
		var userId interface{}
		userIdValue, exists := c.Get("userId")
		if exists {
			userId = userIdValue
		}

		// Loki에 맞는 JSON 형식으로 출력
		jsonLog := fmt.Sprintf(`{"level":"%s","timestamp":%d,"method":"%s","path":"%s","statusCode":%d,"durationMs":%d,"userId":%v}`,
			getLogLevel(statusCode),
			time.Now().Unix(),
			method,
			path,
			statusCode,
			duration.Milliseconds(),
			userId,
		)
		fmt.Println(jsonLog)
	}
}

func getLogLevel(statusCode int) string {
	switch {
	case statusCode >= 500:
		return "error"
	case statusCode >= 400:
		return "warn"
	default:
		return "info"
	}
}
