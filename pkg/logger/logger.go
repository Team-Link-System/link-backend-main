package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"go.uber.org/zap"
)

type Logger struct {
	errorFile   *os.File
	successFile *os.File
	zapLogger   *zap.Logger
}

var logger *Logger

// InitLogger initializes the logger
func InitLogger() error {
	// Create logs directory
	if err := os.MkdirAll("logs", 0755); err != nil {
		return fmt.Errorf("로그 디렉토리 생성 실패: %v", err)
	}

	// Current date for log filenames
	currentDate := time.Now().Format("2006-01-02")
	errorLogPath := filepath.Join("logs", fmt.Sprintf("error_%s.log", currentDate))
	successLogPath := filepath.Join("logs", fmt.Sprintf("success_%s.log", currentDate))

	// Open error log file
	errorFile, err := os.OpenFile(errorLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("에러 로그 파일 생성 실패: %v", err)
	}

	// Open success log file
	successFile, err := os.OpenFile(successLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("성공 로그 파일 생성 실패: %v", err)
	}

	logger = &Logger{
		errorFile:   errorFile,
		successFile: successFile,
	}
	return nil
}

// LogError logs an error message
func LogError(message string) error {
	if logger == nil {
		if err := InitLogger(); err != nil {
			return err
		}
	}

	_, file, line, _ := runtime.Caller(2)
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	jsonLog := fmt.Sprintf(`{"level":"error","timestamp":"%s","file":"%s","line":%d,"message":"%s"}`,
		timestamp, filepath.Base(file), line, message)

	logEntry := fmt.Sprintf("%s\n", jsonLog)
	_, err := logger.errorFile.WriteString(logEntry)

	// stdout에 JSON 형식으로 출력
	fmt.Println(logEntry) // 줄바꿈이 자동으로 추가됨

	return err
}

// LogSuccess logs a success message
func LogSuccess(message string) error {
	if logger == nil {
		if err := InitLogger(); err != nil {
			return err
		}
	}

	_, file, line, _ := runtime.Caller(2)
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	jsonLog := fmt.Sprintf(`{"level":"info","timestamp":"%s","file":"%s","line":%d,"message":"%s"}`,
		timestamp, filepath.Base(file), line, message)
	logEntry := fmt.Sprintf("%s\n", jsonLog)
	_, err := logger.successFile.WriteString(logEntry)

	fmt.Println(logEntry) // 줄바꿈이 자동으로 추가됨

	return err
}

// CloseLogger closes all log files
func CloseLogger() {
	if logger != nil {
		if logger.errorFile != nil {
			logger.errorFile.Close()
		}
		if logger.successFile != nil {
			logger.successFile.Close()
		}
	}
}
