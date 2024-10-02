package interceptor

import (
	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // 요청 처리 진행

		// 에러 발생 시 처리
		err := c.Errors.Last()
		if err != nil {
			// 글로벌 응답 처리와 연동하여 에러 응답 반환
			c.JSON(c.Writer.Status(), Error(c.Writer.Status(), err.Error()))
		}
	}
}
