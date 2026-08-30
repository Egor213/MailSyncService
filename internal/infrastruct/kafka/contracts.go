package kafka

import "context"

type KafkaProducer interface {
	PublishSyncEvent(ctx context.Context, mailboxID string) error
	Close() error
}
