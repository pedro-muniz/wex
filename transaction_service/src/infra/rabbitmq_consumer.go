package infra

import (
	"context"
	"log"

	"wex/transaction_service/src/core/services"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConsumer struct {
	channel *amqp.Channel
	service *services.TransactionPersistenceService
}

func NewRabbitMQConsumer(channel *amqp.Channel, queueName string, service *services.TransactionPersistenceService) *RabbitMQConsumer {
	// queueName is kept for backwards compatibility in DI but not used directly
	return &RabbitMQConsumer{
		channel: channel,
		service: service,
	}
}

func (c *RabbitMQConsumer) Start(ctx context.Context) error {
	if err := c.consumeQueue(ctx, "transaction_jobs", func(ctx context.Context, id uuid.UUID) error {
		log.Printf("Processing job: %s", id)
		return c.service.ProcessTransaction(ctx, id)
	}); err != nil {
		return err
	}

	if err := c.consumeQueue(ctx, "sync_jobs", func(ctx context.Context, id uuid.UUID) error {
		log.Printf("Processing sync job: %s", id)
		return c.service.SyncCache(ctx, id)
	}); err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}

func (c *RabbitMQConsumer) consumeQueue(ctx context.Context, queueName string, handler func(context.Context, uuid.UUID) error) error {
	_, err := c.channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	msgs, err := c.channel.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			id, err := uuid.Parse(string(d.Body))
			if err != nil {
				log.Printf("invalid job id in %s: %s", queueName, d.Body)
				continue
			}

			if err := handler(ctx, id); err != nil {
				log.Printf("failed to process job %s in %s: %v", id, queueName, err)
			}
		}
	}()

	log.Printf("Worker listening on queue: %s", queueName)
	return nil
}
