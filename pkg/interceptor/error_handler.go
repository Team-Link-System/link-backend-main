package interceptor

import (
	"fmt"
	"link/pkg/common"
	"link/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 에러가 있는 경우에만 처리
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			var statusCode int
			var message string

			// AppError 타입 체크
			if appErr, ok := err.Err.(*common.AppError); ok {
				statusCode = appErr.StatusCode
				message = appErr.Message
			} else {
				statusCode = http.StatusInternalServerError
				message = "서버 에러"
			}

			// 에러 로그 기록
			// 에러 로그 기록 (스택 트레이스 포함)
			errorWithStack := errors.WithStack(err.Err) // 스택 트레이스 추가
			errorMsg := fmt.Sprintf(
				"경로: %s, 메소드: %s, 클라이언트 IP: %s, 상태 코드: %d, 에러 메시지: %s\n스택 트레이스:\n%+v",
				c.FullPath(), c.Request.Method, c.ClientIP(), statusCode, err.Error(), errorWithStack,
			)
			logger.LogError(errorMsg)

			// 클라이언트에 응답
			c.JSON(statusCode, gin.H{
				"status":  statusCode,
				"message": message,
			})
			c.Abort()
		}
	}
}
