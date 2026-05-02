//go:build wireinject
// +build wireinject

package di

import (
	"database/sql"

	"github.com/google/wire"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"wex/conversion_service/src/core/ports"
	"wex/conversion_service/src/core/services"
	"wex/conversion_service/src/infra"
	"wex/conversion_service/src/infra/dao"
	"wex/conversion_service/src/infra/repositories"
)

// ServiceSet defines the singleton services for the application.
var ServiceSet = wire.NewSet(
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
)

// DAOSet defines the singleton DAOs.
var DAOSet = wire.NewSet(
	dao.NewPostgresDAO,
	dao.NewTreasuryAPIDAO,
	dao.NewValkeyDAO,
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
