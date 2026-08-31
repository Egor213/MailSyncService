package redis

import (
	"context"
	redispkg "mail-sync-service/pkg/redis"
	"time"
)

type RedisLocker struct {
	client *redispkg.Redis
}

func NewRedisLocker(client *redispkg.Redis) *RedisLocker {
	return &RedisLocker{client: client}
}

func (l *RedisLocker) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return l.client.Lock(ctx, key, ttl)
}

func (l *RedisLocker) Unlock(ctx context.Context, key string) error {
	return l.client.Unlock(ctx, key)
}
