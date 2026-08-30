package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis – обёртка над клиентом Redis.
type Redis struct {
	Client *redis.Client
}

// New создаёт новый клиент Redis и проверяет соединение.
func New(address, password string, db int) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Redis{Client: client}, nil
}

// Close закрывает соединение с Redis.
func (r *Redis) Close() error {
	return r.Client.Close()
}

// --- Базовые операции ---

// Get возвращает значение по ключу.
func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

// Set записывает значение с TTL (0 – без TTL).
func (r *Redis) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.Client.Set(ctx, key, value, ttl).Err()
}

// Del удаляет один или несколько ключей.
func (r *Redis) Del(ctx context.Context, keys ...string) error {
	return r.Client.Del(ctx, keys...).Err()
}

// Exists проверяет существование ключа.
func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.Client.Exists(ctx, key).Result()
	return n > 0, err
}

// --- Распределённые блокировки (на основе SetNX) ---

// Lock пытается захватить блокировку с указанным TTL.
// Возвращает true, если блокировка успешно захвачена, иначе false.
func (r *Redis) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return r.Client.SetNX(ctx, key, "locked", ttl).Result()
}

// Unlock освобождает блокировку.
// В простейшем случае просто удаляет ключ.
// Для атомарности в продакшене стоит использовать Lua-скрипт,
// но для демонстрации оставляем так.
func (r *Redis) Unlock(ctx context.Context, key string) error {
	return r.Client.Del(ctx, key).Err()
}

// --- Дополнительно: атомарная блокировка с Lua-скриптом (опционально) ---

const unlockScript = `
	if redis.call("get", KEYS[1]) == ARGV[1] then
		return redis.call("del", KEYS[1])
	else
		return 0
	end
`

// UnlockWithValue атомарно освобождает блокировку, только если значение совпадает.
// value – случайная строка, переданная при захвате (предотвращает удаление чужих блокировок).
func (r *Redis) UnlockWithValue(ctx context.Context, key, value string) error {
	script := redis.NewScript(unlockScript)
	res, err := script.Run(ctx, r.Client, []string{key}, value).Int64()
	if err != nil {
		return err
	}
	if res == 0 {
		return errors.New("lock not held or value mismatch")
	}
	return nil
}
