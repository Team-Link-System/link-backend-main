package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type Logger struct {
	errorFile   *os.File
	successFile *os.File
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
	logEntry := fmt.Sprintf("[%s] ERROR %s:%d - %s\n", timestamp, file, line, message)
	_, err := logger.errorFile.WriteString(logEntry)
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
	logEntry := fmt.Sprintf("[%s] SUCCESS %s:%d - %s\n", timestamp, file, line, message)
	_, err := logger.successFile.WriteString(logEntry)
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
