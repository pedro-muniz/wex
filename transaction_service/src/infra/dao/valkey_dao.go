package dao

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type ValkeyDAO struct {
	client *redis.Client
}

func NewValkeyDAO(client *redis.Client) *ValkeyDAO {
	return &ValkeyDAO{client: client}
}

func (d *ValkeyDAO) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return d.client.Set(ctx, key, value, expiration).Err()
}

func (d *ValkeyDAO) Get(ctx context.Context, key string) (string, error) {
	return d.client.Get(ctx, key).Result()
}

func (d *ValkeyDAO) Del(ctx context.Context, key string) error {
	return d.client.Del(ctx, key).Err()
}

func (d *ValkeyDAO) SetStatus(ctx context.Context, transactionID string, status string, expiration time.Duration) error {
	return d.client.Set(ctx, "status:"+transactionID, status, expiration).Err()
}

func (d *ValkeyDAO) GetStatus(ctx context.Context, transactionID string) (string, error) {
	return d.client.Get(ctx, "status:"+transactionID).Result()
}
