package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type Logger struct {
	errorFile *os.File
}

var logger *Logger

func InitLogger() error {
	// logs 디렉토리 생성
	if err := os.MkdirAll("logs", 0755); err != nil {
		return fmt.Errorf("로그 디렉토리 생성 실패: %v", err)
	}

	// 현재 날짜로 로그 파일 생성
	currentDate := time.Now().Format("2006-01-02")
	errorLogPath := filepath.Join("logs", fmt.Sprintf("error_%s.log", currentDate))

	errorFile, err := os.OpenFile(errorLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("에러 로그 파일 생성 실패: %v", err)
	}

	logger = &Logger{
		errorFile: errorFile,
	}
	return nil
}

func LogError(message string) error {
	if logger == nil {
		if err := InitLogger(); err != nil {
			return err
		}
	}

	// timestamp := time.Now().Format("2006-01-02 15:04:05")
	// logEntry := fmt.Sprintf("[%s] %s\n", timestamp, message)

	// _, err := logger.errorFile.WriteString(logEntry)
	// return err

	// runtime.Caller를 통해 파일 경로와 라인 번호 얻기
	_, file, line, ok := runtime.Caller(2) // 2는 호출 스택의 깊이
	if ok {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		logEntry := fmt.Sprintf("[%s] %s:%d - %s\n", timestamp, file, line, message)
		_, err := logger.errorFile.WriteString(logEntry)
		return err
	}
	return fmt.Errorf("파일 경로나 라인 번호를 가져오지 못했습니다")
}

func CloseLogger() {
	if logger != nil && logger.errorFile != nil {
		logger.errorFile.Close()
	}
}
