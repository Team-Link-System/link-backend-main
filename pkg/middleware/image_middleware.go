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
		uniqueFileName := uuid.New().String()[:15]
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

// TODO 게시글 이미지 파일 업로드 미들웨어
func (i *ImageUploadMiddleware) PostImageUploadMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		files, err := c.MultipartForm()
		if err != nil {
			c.Next()
			return
		}

		formFiles := files.File["files"]
		if len(formFiles) == 0 {
			c.Next()
			return
		}

		imageUrls := []string{}

		for _, file := range formFiles {
			ext := strings.ToLower(filepath.Ext(file.Filename))
			if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
				c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "허용되지 않는 파일 형식입니다"))
				c.Abort()
				return
			}

			now := time.Now().Format("2006-01-02")
			folderPath := filepath.Join(i.directory, now)

			uniqueFileName := uuid.New().String()[:15]
			fileName := uniqueFileName + ext
			filePath := filepath.Join(folderPath, fileName)

			if err := c.SaveUploadedFile(file, filePath); err != nil {
				c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "파일 저장 실패"))
				c.Abort()
				return
			}

			imageUrls = append(imageUrls, fmt.Sprintf("%s/%s/%s", i.staticPrefix, now, fileName))
		}

		//TODO next로 넘길때 배열 형태로 넘겨주기
		c.Set("image_urls", imageUrls)
		c.Next()
	}
}
