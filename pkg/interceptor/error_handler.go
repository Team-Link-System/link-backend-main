package interceptor

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 에러 미들웨어 (사용 안함)
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // 요청 처리 진행

		err := c.Errors.Last()
		if err != nil {
			statusCode := http.StatusInternalServerError
			if meta, ok := err.Meta.(int); ok {
				statusCode = meta
			}

			log.Printf(
				"[에러] 경로: %s, 메소드: %s, 클라이언트 IP: %s, 상태 코드: %d, 에러 메시지: %s",
				c.FullPath(), c.Request.Method, c.ClientIP(), statusCode, err.Error(),
			)

			c.JSON(statusCode, gin.H{"error": err.Error()})
		}
	}
}
