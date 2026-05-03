package infra

import (
	"context"
	"fmt"
	"log"
	"strings"

	"wex/conversion_service/src/core/ports"
	"wex/conversion_service/src/core/services"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConsumer struct {
	channel      *amqp.Channel
	queueName    string
	service      *services.TransactionQueryService
	payloadStore ports.PayloadStore
}

func NewRabbitMQConsumer(channel *amqp.Channel, queueName string, service *services.TransactionQueryService, payloadStore ports.PayloadStore) *RabbitMQConsumer {
	return &RabbitMQConsumer{
		channel:      channel,
		queueName:    queueName,
		service:      service,
		payloadStore: payloadStore,
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
		return fmt.Errorf("failed to declare queue: %w", err)
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
			body := string(d.Body)
			parts := strings.Split(body, ":")
			if len(parts) != 2 {
				log.Printf("invalid job body: %s", body)
				continue
			}

			id, err := uuid.Parse(parts[0])
			if err != nil {
				log.Printf("invalid uuid: %s", parts[0])
				continue
			}
			currency := parts[1]

			log.Printf("[Worker] Processing conversion request: ID=%s, TargetCurrency=%s", id, currency)

			_, err = c.service.GetConvertedTransaction(ctx, id, currency)
			if err != nil {
				log.Printf("[Worker] [ERROR] Conversion failed for %s: %v", id, err)
				continue
			}
		}
	}()

	log.Printf("Conversion Worker listening on queue: %s", c.queueName)
	<-ctx.Done()
	return nil
}
