package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	HTTPPort string
	WSPath   string
	WSPort   string
	DB       *gorm.DB
	Redis    *redis.Client
	Mongo    *mongo.Client
}

func LoadConfig() *Config {
	LoadEnv()

	return &Config{
		HTTPPort: ":" + getEnv("HTTP_PORT", "8080"),
		WSPort:   ":" + getEnv("WS_PORT", "8081"),
		WSPath:   getEnv("WS_PATH", "/ws"),
		DB:       InitDB(),
		Redis:    InitRedis(),
		Mongo:    InitMongo(),
	}
}

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
		log.Printf("에러 %s 파일 로드", envFile)
	} else {
		log.Printf("%s 파일 로드 성공", envFile)
	}
}

func InitDB() *gorm.DB {
	dsn := os.Getenv("POSTGRES_DSN")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("POSTGRES DB 연결 오류: ", err)
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
		log.Fatal("레디스 연결 오류:", err)
	}

	return rdb
}

func InitMongo() *mongo.Client {

	mongoURI := os.Getenv("MONGO_DSN")
	clientOptions := options.Client().ApplyURI(mongoURI).SetConnectTimeout(10 * time.Second)
	// MongoDB 클라이언트 초기화
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatalf("몽고DB 연결 오류: %v", err)
	}

	// 연결 확인
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatalf("몽고DB 연결 오류: %v", err)
	}

	fmt.Println("몽고DB 연결 성공")
	return client
}

func parseRedisDB(db string) int {
	i, _ := strconv.Atoi(db)
	return i
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func EnsureDirectory(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			log.Fatalf("폴더 생성 실패: %s, 오류: %v", path, err)
		}
	}
}
