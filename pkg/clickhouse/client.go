package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

// Client обёртка над *sql.DB для ClickHouse
type Client struct {
	db *sql.DB
}

// NewClient создаёт новое подключение к ClickHouse.
// dsn пример: "clickhouse://localhost:9000?username=default&password=&database=mail_sync"
func NewClient(dsn string) (*Client, error) {
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return nil, fmt.Errorf("clickhouse open: %w", err)
	}
	// Проверяем подключение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("clickhouse ping: %w", err)
	}
	return &Client{db: db}, nil
}

// Exec выполняет запрос без возврата строк (INSERT, UPDATE, DELETE)
func (c *Client) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return c.db.ExecContext(ctx, query, args...)
}

// Query выполняет запрос с возвратом строк
func (c *Client) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return c.db.QueryContext(ctx, query, args...)
}

// QueryRow выполняет запрос, возвращающий одну строку
func (c *Client) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return c.db.QueryRowContext(ctx, query, args...)
}

// Close закрывает соединение с БД
func (c *Client) Close() error {
	return c.db.Close()
}
