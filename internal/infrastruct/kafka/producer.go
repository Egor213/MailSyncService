// package kafka

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"time"

// 	"github.com/IBM/sarama"
// 	log "github.com/sirupsen/logrus"
// )

// // Producer — реализация KafkaProducer
// type Producer struct {
// 	producer sarama.SyncProducer
// 	topic    string
// }

// // NewProducer создаёт нового Kafka-продюсера
// func NewProducer(brokers []string, topic string) (*Producer, error) {
// 	config := sarama.NewConfig()
// 	config.Producer.RequiredAcks = sarama.WaitForAll          // ждём подтверждения от всех реплик
// 	config.Producer.Retry.Max = 3                             // повторные попытки
// 	config.Producer.Return.Successes = true                   // возвращать успешные отправки
// 	config.Producer.Partitioner = sarama.NewRandomPartitioner // случайная партиция

// 	producer, err := sarama.NewSyncProducer(brokers, config)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
// 	}

// 	return &Producer{
// 		producer: producer,
// 		topic:    topic,
// 	}, nil
// }

// // PublishSyncEvent отправляет событие синхронизации для почтового ящика
// func (p *Producer) PublishSyncEvent(ctx context.Context, mailboxID string) error {
// 	event := struct {
// 		MailboxID string `json:"mailbox_id"`
// 		Type      string `json:"type"`
// 		Timestamp int64  `json:"timestamp"`
// 	}{
// 		MailboxID: mailboxID,
// 		Type:      "mailbox.sync",
// 		Timestamp: time.Now().Unix(),
// 	}

// 	data, err := json.Marshal(event)
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal event: %w", err)
// 	}

// 	msg := &sarama.ProducerMessage{
// 		Topic: p.topic,
// 		Key:   sarama.StringEncoder(mailboxID), // ключ для гарантии порядка сообщений для одного ящика
// 		Value: sarama.ByteEncoder(data),
// 	}

// 	partition, offset, err := p.producer.SendMessage(msg)
// 	if err != nil {
// 		return fmt.Errorf("failed to send message: %w", err)
// 	}

// 	log.WithFields(log.Fields{
// 		"mailbox_id": mailboxID,
// 		"partition":  partition,
// 		"offset":     offset,
// 		"topic":      p.topic,
// 	}).Debug("kafka event published")

// 	return nil
// }

// // Close закрывает продюсера
// func (p *Producer) Close() error {
// 	return p.producer.Close()
// }
