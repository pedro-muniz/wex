//go:build wireinject
// +build wireinject

package di

import (
	"database/sql"

	"github.com/google/wire"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"wex/api_service/src/controllers"
	"wex/api_service/src/core/ports"
	"wex/api_service/src/core/services"
	"wex/api_service/src/infra/dao"
	"wex/api_service/src/infra/repositories"
)

// ServiceSet defines the singleton services for the application.
var ServiceSet = wire.NewSet(
	services.NewTransactionProducerService,
	services.NewConversionProducerService,
	services.NewTransactionQueryService,
)

// RepoSet defines the singleton repositories.
var RepoSet = wire.NewSet(
	repositories.NewTransactionRepository,
	wire.Bind(new(ports.TransactionRepository), new(*repositories.TransactionRepository)),

	repositories.NewTreasuryRateProvider,
	wire.Bind(new(ports.ConversionRateProvider), new(*repositories.TreasuryRateProvider)),

	repositories.NewValkeyPayloadStore,
	wire.Bind(new(ports.PayloadStore), new(*repositories.ValkeyPayloadStore)),

	repositories.NewRabbitMQPublisher,
	wire.Bind(new(ports.MessagePublisher), new(*repositories.RabbitMQPublisher)),
)

// DAOSet defines the singleton DAOs.
var DAOSet = wire.NewSet(
	dao.NewPostgresDAO,
	dao.NewValkeyDAO,
	dao.NewRabbitMQDAO,
	dao.NewTreasuryAPIDAO,
	
	// Internal interface bindings for repositories
	wire.Bind(new(repositories.RabbitMQPublisherDAO), new(*dao.RabbitMQDAO)),
	wire.Bind(new(repositories.ValkeyDAO), new(*dao.ValkeyDAO)),
	wire.Bind(new(repositories.PostgresDAO), new(*dao.PostgresDAO)),
	wire.Bind(new(repositories.TreasuryAPIDAO), new(*dao.TreasuryAPIDAO)),
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
