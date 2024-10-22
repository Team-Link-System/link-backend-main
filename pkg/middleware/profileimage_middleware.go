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

type ImageUploadMiddleware struct {
	directory    string
	staticPrefix string // URL 경로의 Prefix
}

// NewImageUploadMiddleware는 ImageUploadMiddleware를 생성하는 함수입니다.
func NewImageUploadMiddleware(directory, staticPrefix string) *ImageUploadMiddleware {
	return &ImageUploadMiddleware{
		directory:    directory,
		staticPrefix: staticPrefix,
	}
}

// ProfileImageUploadMiddleware는 이미지 업로드를 처리하는 미들웨어 함수입니다.
func (i *ImageUploadMiddleware) ProfileImageUploadMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.Next() // 이미지가 없으면 다음 핸들러로
			return
		}

		ext := strings.ToLower(filepath.Ext(file.Filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
			c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "허용되지 않는 파일 형식입니다"))
			c.Abort()
			return
		}

		// 현재 날짜를 기반으로 폴더 경로 생성
		now := time.Now().Format("2006-01-02") // 날짜 포맷 수정
		folderPath := filepath.Join(i.directory, now)

		// 폴더가 없으면 생성
		if _, err := os.Stat(folderPath); os.IsNotExist(err) {
			if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
				c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "폴더 생성 실패"))
				c.Abort()
				return
			}
		}

		// 파일 저장 경로 설정
		uniqueFileName := uuid.New().String()
		fileName := uniqueFileName + ext
		filePath := filepath.Join(folderPath, fileName)

		// 원본 파일 저장
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "파일 저장 실패"))
			c.Abort()
			return
		}

		// 이미지 URL 설정
		imageUrl := fmt.Sprintf("%s/%s/%s", i.staticPrefix, now, fileName)
		c.Set("profile_image_url", imageUrl)
		c.Next()
	}
}
