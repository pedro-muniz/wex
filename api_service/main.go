package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "wex/api_service/docs" // Swagger docs
	"wex/api_service/src/infra/di"
	"wex/api_service/src/infra/repositories"
)

// @title Purchase Transaction API
// @version 1.0
// @description API for managing purchase transactions with multi-currency support.
// @host localhost:8080
// @BasePath /
func main() {
	godotenv.Load()

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
	queue := repositories.RabbitMQQueue("transaction_jobs")
	exchange := repositories.RabbitMQExchange("")
	
	apiControllers := di.InitializeAPI(db, ch, redisClient, queue, exchange)

	// Router
	mux := http.NewServeMux()
	apiControllers.RegisterAll(mux)

	// Swagger
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	log.Println("API Gateway Service starting on :8080")
	log.Println("Swagger documentation available at http://localhost:8080/swagger/index.html")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
