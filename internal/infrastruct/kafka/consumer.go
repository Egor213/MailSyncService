package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/IBM/sarama"
	log "github.com/sirupsen/logrus"
)

type Consumer struct {
	consumer sarama.ConsumerGroup
	topic    string
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

type SyncHandler interface {
	HandleSync(ctx context.Context, mailboxID string) error
}

func NewConsumer(brokers []string, groupID, topic string) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("create consumer group: %w", err)
	}
	return &Consumer{
		consumer: consumer,
		topic:    topic,
	}, nil
}

func (c *Consumer) Run(ctx context.Context, handler SyncHandler) error {
	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	h := &syncGroupHandler{handler: handler}
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			if err := c.consumer.Consume(ctx, []string{c.topic}, h); err != nil {
				log.Errorf("consumer error: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()
	return nil
}

func (c *Consumer) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
	c.wg.Wait()
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}

type syncGroupHandler struct {
	handler SyncHandler
}

func (h *syncGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *syncGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *syncGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var event struct {
			MailboxID string `json:"mailbox_id"`
		}
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Errorf("unmarshal event: %v", err)
			continue
		}
		if err := h.handler.HandleSync(sess.Context(), event.MailboxID); err != nil {
			log.WithError(err).WithField("mailbox_id", event.MailboxID).Error("sync handling failed")
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}
