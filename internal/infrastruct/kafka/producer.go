package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	log "github.com/sirupsen/logrus"
)

type Producer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewProducer(brokers []string, topic string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3
	config.Producer.Return.Successes = true
	config.Producer.Partitioner = sarama.NewRandomPartitioner

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}
	return &Producer{producer: producer, topic: topic}, nil
}

func (p *Producer) PublishSyncEvent(ctx context.Context, mailboxID string) error {
	event := struct {
		MailboxID string `json:"mailbox_id"`
		Type      string `json:"type"`
		Timestamp int64  `json:"timestamp"`
	}{
		MailboxID: mailboxID,
		Type:      "mailbox.sync",
		Timestamp: time.Now().Unix(),
	}
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(mailboxID),
		Value: sarama.ByteEncoder(data),
	}
	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	log.WithFields(log.Fields{
		"mailbox_id": mailboxID,
		"partition":  partition,
		"offset":     offset,
		"topic":      p.topic,
	}).Debug("kafka event published")
	return nil
}

func (p *Producer) Close() error {
	return p.producer.Close()
}
