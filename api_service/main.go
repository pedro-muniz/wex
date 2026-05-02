package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

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
