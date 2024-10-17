package interceptor

import (
	"log"
	"net/http"

	"link/pkg/common"

	"github.com/gin-gonic/gin"
)

// 에러 미들웨어 (사용 안함)
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // 요청 처리 진행
		err := c.Errors.Last()
		if err != nil {
			statusCode, ok := err.Meta.(int)
			if !ok {
				statusCode = http.StatusInternalServerError
			}

			// 에러 응답을 Response 구조체로 처리하여 JSON으로 반환
			log.Printf("에러 발생: %v", err.Error())
			c.JSON(statusCode, common.NewError(statusCode, err.Error()))
		}
	}
}
