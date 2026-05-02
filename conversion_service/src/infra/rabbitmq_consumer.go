package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"wex/conversion_service/src/core/ports"
	"wex/conversion_service/src/core/services"
)

type RabbitMQConsumer struct {
	channel    *amqp.Channel
	queueName  string
	service    *services.TransactionQueryService
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

			log.Printf("Processing conversion for: %s to %s", id, currency)
			
			resp, err := c.service.GetConvertedTransaction(ctx, id, currency)
			if err != nil {
				log.Printf("conversion error: %v", err)
				continue
			}

			respData, _ := json.Marshal(resp)
			valkeyKey := fmt.Sprintf("conversion:%s:%s", id, currency)
			
			if err := c.payloadStore.SetRaw(ctx, valkeyKey, string(respData)); err != nil {
				log.Printf("failed to store result in valkey: %v", err)
			}
		}
	}()

	log.Printf("Conversion Worker listening on queue: %s", c.queueName)
	<-ctx.Done()
	return nil
}
