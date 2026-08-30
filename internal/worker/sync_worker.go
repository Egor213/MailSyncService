package worker

import (
	"context"
	"encoding/json"
	"mail-sync-service/internal/service"

	"github.com/IBM/sarama"
	log "github.com/sirupsen/logrus"
)

type SyncWorker struct {
	consumer    sarama.ConsumerGroup
	syncService *service.SyncService
	topic       string
}

func (w *SyncWorker) Run(ctx context.Context) error {
	handler := &syncHandler{syncService: w.syncService}
	for {
		if err := w.consumer.Consume(ctx, []string{w.topic}, handler); err != nil {
			log.Errorf("consumer error: %v", err)
			return err
		}
	}
}

type syncHandler struct {
	syncService *service.SyncService
}

func (h *syncHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *syncHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *syncHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var event struct {
			MailboxID string `json:"mailbox_id"`
		}
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Errorf("unmarshal error: %v", err)
			continue
		}
		if err := h.syncService.SyncMailbox(sess.Context(), event.MailboxID); err != nil {
			log.Errorf("sync error: %v", err)
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}
