package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"wex/conversion_service/src/infra/di"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Infrastructure Setup
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rabbitConn, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitConn.Close()

	ch, err := rabbitConn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("VALKEY_URL"),
	})

	// DI Initialization
	queueName := "conversion_jobs"
	worker := di.InitializeWorker(db, ch, redisClient, queueName)

	log.Println("Conversion Worker Service starting...")
	if err := worker.Start(ctx); err != nil {
		log.Printf("Worker stopped: %v", err)
	}
}
