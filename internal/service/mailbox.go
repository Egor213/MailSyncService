package service

import (
	"context"
	"mail-sync-service/internal/entity"
	"mail-sync-service/internal/infrastruct/kafka"
	"mail-sync-service/internal/repo"
	"time"
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

// CreateMailboxInput – DTO для создания ящика (принимает строковые названия)
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
	return nil, nil
}

func (s *MailboxService) GetMailbox(ctx context.Context, id string) (*entity.Mailbox, error) {
	return nil, nil
}

func (s *MailboxService) UpdateMailbox(ctx context.Context, mb *entity.Mailbox) error {
	return nil
}

func (s *MailboxService) DeleteMailbox(ctx context.Context, id string) error {
	return nil
}

func (s *MailboxService) ListActive(ctx context.Context) ([]*entity.Mailbox, error) {
	return nil, nil
}

func (s *MailboxService) TriggerSync(ctx context.Context, mailboxID string) error {
	return nil
}

// func (s *MailboxService) CreateMailbox(ctx context.Context, in CreateMailboxInput) (*entity.Mailbox, error) {
// 	// Получаем ID справочников
// 	providerID, err := s.refRepo.GetProviderID(ctx, in.Provider)
// 	if err != nil {
// 		return nil, errors.New("unknown provider: " + in.Provider)
// 	}
// 	protocolID, err := s.refRepo.GetProtocolID(ctx, in.Protocol)
// 	if err != nil {
// 		return nil, errors.New("unknown protocol: " + in.Protocol)
// 	}
// 	authTypeID, err := s.refRepo.GetAuthTypeID(ctx, in.AuthType)
// 	if err != nil {
// 		return nil, errors.New("unknown auth type: " + in.AuthType)
// 	}

// 	mb := &entity.Mailbox{
// 		ID:           uuid.New().String(),
// 		Email:        in.Email,
// 		ProviderID:   providerID,
// 		ProtocolID:   protocolID,
// 		Server:       in.Server,
// 		Port:         in.Port,
// 		UseTLS:       in.UseTLS,
// 		AuthTypeID:   authTypeID,
// 		AccessToken:  in.AccessToken,
// 		RefreshToken: in.RefreshToken,
// 		TokenExpiry:  in.TokenExpiry,
// 		CreatedAt:    time.Now(),
// 		UpdatedAt:    time.Now(),
// 		IsActive:     true,
// 	}

// 	if err := s.mailboxRepo.Create(ctx, mb); err != nil {
// 		return nil, err
// 	}

// 	// Отправляем событие в Kafka
// 	_ = s.producer.PublishSyncEvent(ctx, mb.ID)
// 	return mb, nil
// }
