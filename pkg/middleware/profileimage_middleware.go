package middleware

import (
	"fmt"
	"link/pkg/common"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ProfileImageUploadMiddleware(directory string) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("image")
		if err != nil {
			c.Next() //이미지 없으면 다음 핸들러
			return
		}

		ext := strings.ToLower(filepath.Ext(file.Filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
			c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "허용되지 않는 파일 형식입니다"))
			c.Abort()
			return
		}

		//TODO 리사이징 처리(이후)

		//현재 날짜를 기반으로 폴더 경로 생성
		now := time.Now().Format("2024-01-01")
		folderPath := filepath.Join(directory, now)

		//없으면 생성
		if _, err := os.Stat(folderPath); os.IsNotExist(err) {
			if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
				c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "폴더 생성 실패"))
				c.Abort()
				return
			}
		}

		uniqueFileName := uuid.New().String() + ext
		savePath := filepath.Join(folderPath, uniqueFileName)

		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "파일 저장 실패"))
			c.Abort()
			return
		}

		imageUrl := fmt.Sprintf("/static/profiles/%s/%s", now, uniqueFileName)
		c.Set("profile_image_url", imageUrl)
		c.Next()
	}
}
