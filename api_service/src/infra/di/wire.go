//go:build wireinject
// +build wireinject

package di

import (
	"database/sql"

	"wex/api_service/src/controllers"
	"wex/api_service/src/core/ports"
	"wex/api_service/src/core/services"
	"wex/api_service/src/infra/dao"
	"wex/api_service/src/infra/repositories"

	"github.com/google/wire"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

// ServiceSet defines the singleton services for the application.
var ServiceSet = wire.NewSet(
	services.NewTransactionProducerService,
	services.NewConversionProducerService,
)

// RepoSet defines the singleton repositories.
var RepoSet = wire.NewSet(
	repositories.NewValkeyPayloadStore,
	wire.Bind(new(ports.PayloadStore), new(*repositories.ValkeyPayloadStore)),

	repositories.NewRabbitMQPublisher,
	wire.Bind(new(ports.MessagePublisher), new(*repositories.RabbitMQPublisher)),
)

// DAOSet defines the singleton DAOs.
var DAOSet = wire.NewSet(
	dao.NewValkeyDAO,
	dao.NewRabbitMQDAO,

	// Internal interface bindings for repositories
	wire.Bind(new(repositories.RabbitMQPublisherDAO), new(*dao.RabbitMQDAO)),
	wire.Bind(new(repositories.ValkeyDAO), new(*dao.ValkeyDAO)),
)

func InitializeAPI(
	db *sql.DB,
	amqpChannel *amqp.Channel,
	redisClient *redis.Client,
	queue repositories.RabbitMQQueue,
	exchange repositories.RabbitMQExchange,
) *controllers.APIControllers {
	wire.Build(
		DAOSet,
		RepoSet,
		ServiceSet,
		controllers.NewTransactionController,
		controllers.NewConversionController,
		controllers.NewAPIControllers,
	)
	return &controllers.APIControllers{}
}
