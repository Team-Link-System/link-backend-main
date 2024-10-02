package config

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func LoadEnv() {
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "dev" // 기본값을 개발 환경으로 설정
	}

	envFile := ".env"
	if env != "prod" {
		envFile = ".env." + env
	}

	err := godotenv.Load(envFile)
	if err != nil {
		log.Printf("Error loading %s file", envFile)
	} else {
		log.Printf("Loaded %s file", envFile)
	}
}

func InitDB() *gorm.DB {
	dsn := os.Getenv("POSTGRES_DSN")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("POSTGRES DATABASE CONNECTION ERROR: ", err)
	}

	// 개발자 모드에서 디버그 모드 활성화
	if os.Getenv("GO_ENV") == "dev" {
		db = db.Debug()
	}
	return db
}

func InitRedis() *redis.Client {
	redisAddr := os.Getenv("REDIS_ADDR")
	redisDB := os.Getenv("REDIS_DB")

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   parseRedisDB(redisDB),
	})

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	return rdb
}

func parseRedisDB(db string) int {
	i, _ := strconv.Atoi(db)
	return i
}
