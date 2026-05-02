package repositories

import (
	"context"
	"testing"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRabbitMQDAO struct {
	mock.Mock
}

func (m *MockRabbitMQDAO) Publish(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	args := m.Called(ctx, exchange, key, mandatory, immediate, msg)
	return args.Error(0)
}

func TestRabbitMQPublisher(t *testing.T) {
	mockDAO := new(MockRabbitMQDAO)
	publisher := NewRabbitMQPublisher(mockDAO, "test_queue", "test_exchange")
	ctx := context.Background()
	id := uuid.New()

	t.Run("PublishJob", func(t *testing.T) {
		mockDAO.On("Publish", ctx, "test_exchange", "test_queue", false, false, mock.MatchedBy(func(p amqp.Publishing) bool {
			return string(p.Body) == id.String()
		})).Return(nil)

		err := publisher.PublishJob(ctx, id)
		assert.NoError(t, err)
		mockDAO.AssertExpectations(t)
	})

	t.Run("PublishConversionRequest", func(t *testing.T) {
		mockDAO.On("Publish", ctx, "test_exchange", "conversion_jobs", false, false, mock.MatchedBy(func(p amqp.Publishing) bool {
			return string(p.Body) == id.String()+":BRL"
		})).Return(nil)

		err := publisher.PublishConversionRequest(ctx, id, "BRL")
		assert.NoError(t, err)
	})
}
