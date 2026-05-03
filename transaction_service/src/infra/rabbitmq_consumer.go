package infra

import (
	"context"
	"log"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"wex/transaction_service/src/core/services"
)

type RabbitMQConsumer struct {
	channel    *amqp.Channel
	queueName  string
	service    *services.TransactionPersistenceService
}

func NewRabbitMQConsumer(channel *amqp.Channel, queueName string, service *services.TransactionPersistenceService) *RabbitMQConsumer {
	return &RabbitMQConsumer{
		channel:   channel,
		queueName: queueName,
		service:   service,
	}
}

func (c *RabbitMQConsumer) Start(ctx context.Context) error {
	_, err := c.channel.QueueDeclare(
		c.queueName,
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
		c.queueName,
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
				log.Printf("invalid job id: %s", d.Body)
				continue
			}

			log.Printf("Processing job: %s", id)
			if err := c.service.ProcessTransaction(ctx, id); err != nil {
				log.Printf("failed to process job %s: %v", id, err)
			}
		}
	}()

	log.Printf("Worker listening on queue: %s", c.queueName)
	<-ctx.Done()
	return nil
}
