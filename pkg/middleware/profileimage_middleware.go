package middleware

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"link/pkg/common"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anthonynsimon/bild/transform"
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

// UploadMiddleware는 이미지 업로드를 처리하는 미들웨어 함수입니다.
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
		// 파일 저장 경로 설정
		uniqueFileName := uuid.New().String()
		originalFileName := uniqueFileName + "-original" + ext
		originalFilePath := filepath.Join(folderPath, originalFileName)

		// 원본 파일 저장
		if err := c.SaveUploadedFile(file, originalFilePath); err != nil {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "파일 저장 실패"))
			c.Abort()
			return
		}

		// 이미지 파일 열기
		srcFile, err := os.Open(originalFilePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "파일 열기 실패"))
			c.Abort()
			return
		}
		defer srcFile.Close()

		var img image.Image
		switch ext {
		case ".jpg", ".jpeg":
			img, err = jpeg.Decode(srcFile)
		case ".png":
			img, err = png.Decode(srcFile)
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "이미지 디코딩 실패"))
			c.Abort()
			return
		}

		// 이미지 리사이징 (300x300으로 리사이징)
		resizedImg := transform.Resize(img, 300, 300, transform.Linear)

		// 리사이징된 이미지 저장
		resizedFileName := uniqueFileName + ext
		resizedFilePath := filepath.Join(folderPath, resizedFileName)
		outFile, err := os.Create(resizedFilePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "리사이즈된 파일 생성 실패"))
			c.Abort()
			return
		}
		defer outFile.Close()

		switch ext {
		case ".jpg", ".jpeg":
			err = jpeg.Encode(outFile, resizedImg, nil)
		case ".png":
			err = png.Encode(outFile, resizedImg)
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "리사이즈된 이미지 저장 실패"))
			c.Abort()
			return
		}

		// 이미지 URL 설정
		imageUrl := fmt.Sprintf("%s/%s/%s", i.staticPrefix, now, uniqueFileName)
		c.Set("profile_image_url", imageUrl)
		c.Next()
	}
}
