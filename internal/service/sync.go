package service

import (
	"context"
	"mail-sync-service/internal/entity"
	"mail-sync-service/internal/repo"
	"mail-sync-service/internal/repo/redis"
)

type SyncService struct {
	mailboxRepo repo.Mailbox
	msgRepo     repo.Message
	syncJobRepo repo.SyncJob
	locker      redis.RedisLocker
	refRepo     repo.Reference
}

func NewSyncService(
	mailRepo repo.Mailbox,
	msgRepo repo.Message,
	syncJobRepo repo.SyncJob,
	refRepo repo.Reference,
	l redis.RedisLocker,
) *SyncService {
	return &SyncService{
		mailboxRepo: mailRepo,
		msgRepo:     msgRepo,
		syncJobRepo: syncJobRepo,
		refRepo:     refRepo,
		locker:      l,
	}
}

func (s *SyncService) SyncMailbox(ctx context.Context, mailboxID string) error {
	return nil
}

func (s *SyncService) SyncAllActive(ctx context.Context) error {
	return nil
}

func (s *SyncService) GetLastSyncStatus(ctx context.Context, mailboxID string) (*entity.SyncJob, error) {
	return nil, nil
}

// func (s *SyncService) SyncMailbox(ctx context.Context, mailboxID string) error {
// 	// ... lock ...

// 	mb, err := s.mailboxRepo.GetByID(ctx, mailboxID)
// 	if err != nil {
// 		return err
// 	}

// 	// Получаем ID статусов
// 	statusRunning, _ := s.refRepo.GetSyncStatusID(ctx, "running")
// 	statusSuccess, _ := s.refRepo.GetSyncStatusID(ctx, "success")
// 	statusFailed, _ := s.refRepo.GetSyncStatusID(ctx, "failed")

// 	job := &entity.SyncJob{
// 		ID:        uuid.New().String(),
// 		MailboxID: mailboxID,
// 		StatusID:  statusRunning,
// 		StartedAt: ptrTime(time.Now()),
// 	}
// 	_ = s.syncJobRepo.Create(ctx, job)

// 	var count int
// 	var syncErr error
// 	// ... синхронизация (используем mb.ProviderID и mb.ProtocolID для выбора клиента) ...

// 	now := time.Now()
// 	job.FinishedAt = &now
// 	if syncErr != nil {
// 		job.StatusID = statusFailed
// 		job.ErrorMsg = syncErr.Error()
// 	} else {
// 		job.StatusID = statusSuccess
// 		job.MessagesCount = count
// 	}
// 	_ = s.syncJobRepo.Update(ctx, job)

// 	return syncErr
// }
