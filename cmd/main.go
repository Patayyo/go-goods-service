package main

import (
	"database/sql"
	"fmt"
	"go-test/internal/handler"
	"go-test/internal/logger"
	"go-test/internal/repo"
	"go-test/internal/service"
	"go-test/internal/utils"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
)

func initEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment")
	}
}

func initPostgres() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"))

	db, err := utils.RetryConnectToPostgres(psqlInfo, 10, 2*time.Second)
	if err != nil {
		log.Fatalf("failed to connect to Postgres: %v", err)
	}
	return db
}

func initNATS() *nats.Conn {
	natsConn, err := utils.RetryConnectToNATS(os.Getenv("NATS_URL"), 10, 2*time.Second)
	if err != nil {
		log.Fatalf("failed to connect to NATS: %v", err)
	}
	return natsConn
}

func initRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
}

func initLogger() logger.Logger {
	l, err := logger.NewNatsLogger(os.Getenv("NATS_URL"), os.Getenv("NATS_LOG_TOPIC"))
	if err != nil {
		log.Fatalf("failed to initialize NATS logger: %v", err)
	}
	return l
}

func runServer(handler *handler.GoodHandler) {
	r := gin.Default()
	handler.Router(r)

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = ":8080"
	}

	log.Println("server starting on", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func main() {
	initEnv()

	db := initPostgres()
	defer db.Close()

	natsConn := initNATS()
	defer natsConn.Close()

	redisClient := initRedis()
	logSvc := initLogger()

	repo := repo.NewGoodRepo(db)
	svc := service.NewGoodService(repo, redisClient, logSvc)
	handler := handler.NewGoodHandler(svc)

	runServer(handler)
}
