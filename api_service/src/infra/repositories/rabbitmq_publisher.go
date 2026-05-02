package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQPublisherDAO interface {
	Publish(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
}

type RabbitMQQueue string
type RabbitMQExchange string

type RabbitMQPublisher struct {
	dao      RabbitMQPublisherDAO
	queue    RabbitMQQueue
	exchange RabbitMQExchange
}

func NewRabbitMQPublisher(dao RabbitMQPublisherDAO, queue RabbitMQQueue, exchange RabbitMQExchange) *RabbitMQPublisher {
	return &RabbitMQPublisher{
		dao:      dao,
		queue:    queue,
		exchange: exchange,
	}
}

func (p *RabbitMQPublisher) PublishJob(ctx context.Context, jobID uuid.UUID) error {
	msg := amqp.Publishing{
		ContentType:  "text/plain",
		Body:         []byte(jobID.String()),
		DeliveryMode: amqp.Persistent,
	}
	return p.dao.Publish(
		ctx,
		string(p.exchange),
		string(p.queue),
		false,
		false,
		msg,
	)
}

func (p *RabbitMQPublisher) PublishConversionRequest(ctx context.Context, jobID uuid.UUID, currency string) error {
	msg := amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(fmt.Sprintf("%s:%s", jobID.String(), currency)),
	}
	// We'll use a specific queue for conversions
	return p.dao.Publish(
		ctx,
		string(p.exchange),
		"conversion_jobs",
		false,
		false,
		msg,
	)
}
