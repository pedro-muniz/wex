package dao

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQDAO struct {
	channel *amqp.Channel
}

func NewRabbitMQDAO(channel *amqp.Channel) *RabbitMQDAO {
	return &RabbitMQDAO{channel: channel}
}

func (d *RabbitMQDAO) Publish(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	return d.channel.PublishWithContext(ctx, exchange, key, mandatory, immediate, msg)
}
