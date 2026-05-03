package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"wex/transaction_service/src/infra/di"
)

func main() {
	godotenv.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Infrastructure Setup
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			os.Getenv("POSTGRES_USER"),
			os.Getenv("POSTGRES_PASSWORD"),
			os.Getenv("POSTGRES_HOST"),
			os.Getenv("POSTGRES_PORT"),
			os.Getenv("POSTGRES_DB"),
		)
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	amqpURL := os.Getenv("RABBITMQ_URL")
	if amqpURL == "" {
		amqpURL = fmt.Sprintf("amqp://%s:%s@%s:%s/",
			os.Getenv("RABBITMQ_USER"),
			os.Getenv("RABBITMQ_PASSWORD"),
			os.Getenv("RABBITMQ_HOST"),
			os.Getenv("RABBITMQ_PORT"),
		)
	}
	rabbitConn, err := amqp.Dial(amqpURL)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitConn.Close()

	ch, err := rabbitConn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	redisAddr := os.Getenv("VALKEY_URL")
	if redisAddr == "" {
		redisAddr = fmt.Sprintf("%s:%s",
			os.Getenv("VALKEY_HOST"),
			os.Getenv("VALKEY_PORT"),
		)
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// DI Initialization
	queueName := "transaction_jobs"
	worker := di.InitializeWorker(db, ch, redisClient, queueName)

	log.Println("Transaction Worker Service starting...")
	if err := worker.Start(ctx); err != nil {
		log.Printf("Worker stopped: %v", err)
	}
}
