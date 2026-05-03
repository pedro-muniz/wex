//go:build wireinject
// +build wireinject

package di

import (
	"database/sql"

	"wex/conversion_service/src/core/ports"
	"wex/conversion_service/src/core/services"
	"wex/conversion_service/src/infra"
	"wex/conversion_service/src/infra/dao"
	"wex/conversion_service/src/infra/providers"
	"wex/conversion_service/src/infra/repositories"

	"github.com/google/wire"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

// ServiceSet defines the singleton services for the application.
var ServiceSet = wire.NewSet(
	services.NewConversionService,
)

// RepoSet defines the singleton repositories.
var RepoSet = wire.NewSet(
	repositories.NewTransactionRepository,
	wire.Bind(new(ports.TransactionRepository), new(*repositories.TransactionRepository)),

	repositories.NewTreasuryRateRepository,
	repositories.NewRatePostgresRepository,
	providers.NewRateCacheProvider,
	wire.Bind(new(ports.ConversionRateProvider), new(*providers.RateCacheProvider)),

	repositories.NewValkeyRepository,
	wire.Bind(new(ports.PayloadStore), new(*repositories.ValkeyRepository)),
)

// DAOSet defines the singleton DAOs.
var DAOSet = wire.NewSet(
	dao.NewPostgresDAO,
	wire.Bind(new(repositories.RatePostgresDAO), new(*dao.PostgresDAO)),
	wire.Bind(new(repositories.PostgresDAO), new(*dao.PostgresDAO)),
	
	dao.NewTreasuryAPIDAO,
	wire.Bind(new(repositories.TreasuryAPIDAO), new(*dao.TreasuryAPIDAO)),
	
	dao.NewValkeyDAO,
	wire.Bind(new(repositories.ValkeyDAO), new(*dao.ValkeyDAO)),
)

func InitializeWorker(
	db *sql.DB,
	amqpChannel *amqp.Channel,
	redisClient *redis.Client,
	queueName string,
) *infra.RabbitMQConsumer {
	wire.Build(
		DAOSet,
		RepoSet,
		ServiceSet,
		infra.NewRabbitMQConsumer,
	)
	return &infra.RabbitMQConsumer{}
}
