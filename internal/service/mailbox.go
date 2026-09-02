package service

import (
	"context"
	"errors"
	"mail-sync-service/internal/entity"
	"mail-sync-service/internal/infrastruct/kafka"
	"mail-sync-service/internal/repo"
	"time"

	"github.com/google/uuid"
)

type MailboxService struct {
	mailboxRepo repo.Mailbox
	refRepo     repo.Reference
	producer    kafka.KafkaProducer
}

func NewMailboxService(
	mailboxRepo repo.Mailbox,
	refRepo repo.Reference,
	producer kafka.KafkaProducer,
) *MailboxService {
	return &MailboxService{
		mailboxRepo: mailboxRepo,
		refRepo:     refRepo,
		producer:    producer,
	}
}

type CreateMailboxInput struct {
	Email        string
	Provider     string // "gmail", "outlook" и т.д.
	Protocol     string // "imap", "pop3"
	Server       string
	Port         int
	UseTLS       bool
	AuthType     string // "plain", "oauth2"
	AccessToken  string
	RefreshToken string
	TokenExpiry  *time.Time
}

func (s *MailboxService) CreateMailbox(ctx context.Context, in CreateMailboxInput) (*entity.Mailbox, error) {

	providerID, err := s.refRepo.GetProviderID(ctx, in.Provider)
	if err != nil {
		return nil, errors.New("unknown provider: " + in.Provider)
	}
	protocolID, err := s.refRepo.GetProtocolID(ctx, in.Protocol)
	if err != nil {
		return nil, errors.New("unknown protocol: " + in.Protocol)
	}
	authTypeID, err := s.refRepo.GetAuthTypeID(ctx, in.AuthType)
	if err != nil {
		return nil, errors.New("unknown auth type: " + in.AuthType)
	}

	now := time.Now()
	mb := &entity.Mailbox{
		ID:           uuid.New().String(),
		Email:        in.Email,
		ProviderID:   providerID,
		ProtocolID:   protocolID,
		Server:       in.Server,
		Port:         in.Port,
		UseTLS:       in.UseTLS,
		AuthTypeID:   authTypeID,
		AccessToken:  in.AccessToken,
		RefreshToken: in.RefreshToken,
		TokenExpiry:  in.TokenExpiry,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
	}

	if err := s.mailboxRepo.Create(ctx, mb); err != nil {
		return nil, err
	}

	// Отправляем событие в Kafka
	// if err := s.producer.PublishSyncEvent(ctx, mb.ID); err != nil {
	//     // логируем, но не прерываем создание
	//     // log.WithError(err).Warn("failed to publish sync event")
	// }

	return mb, nil
}

func (s *MailboxService) GetMailbox(ctx context.Context, id string) (*entity.Mailbox, error) {
	return s.mailboxRepo.GetByID(ctx, id)
}

func (s *MailboxService) UpdateMailbox(ctx context.Context, mb *entity.Mailbox) error {
	mb.UpdatedAt = time.Now()
	return s.mailboxRepo.Update(ctx, mb)
}

func (s *MailboxService) DeleteMailbox(ctx context.Context, id string) error {
	return s.mailboxRepo.Delete(ctx, id)
}

func (s *MailboxService) ListActive(ctx context.Context) ([]*entity.Mailbox, error) {
	return s.mailboxRepo.ListActive(ctx)
}

func (s *MailboxService) TriggerSync(ctx context.Context, mailboxID string) error {
	return s.producer.PublishSyncEvent(ctx, mailboxID)
}
