package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"mail-sync-service/internal/entity"
	"mail-sync-service/internal/infrastruct/mail"
	"mail-sync-service/internal/repo"
	repoerrs "mail-sync-service/internal/repo/errors"
	"mail-sync-service/internal/repo/redis"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type SyncService struct {
	mailboxRepo  repo.Mailbox
	msgRepo      repo.Message
	syncJobRepo  repo.SyncJob
	refRepo      repo.Reference
	locker       redis.Locker
	oauthService OAuth
}

func NewSyncService(
	mailRepo repo.Mailbox,
	msgRepo repo.Message,
	syncJobRepo repo.SyncJob,
	refRepo repo.Reference,
	l redis.Locker,
	oauthService OAuth,
) *SyncService {
	return &SyncService{
		mailboxRepo:  mailRepo,
		msgRepo:      msgRepo,
		syncJobRepo:  syncJobRepo,
		refRepo:      refRepo,
		locker:       l,
		oauthService: oauthService,
	}
}

func (s *SyncService) SyncAllActive(ctx context.Context) error {
	mailboxes, err := s.mailboxRepo.ListActive(ctx)
	if err != nil {
		return fmt.Errorf("list active mailboxes: %w", err)
	}
	for _, mb := range mailboxes {
		go func(mailboxID string) {
			if err := s.SyncMailbox(context.Background(), mailboxID); err != nil {
				log.WithError(err).WithField("mailbox_id", mailboxID).Error("sync failed")
			}
		}(mb.ID)
	}
	return nil
}

func (s *SyncService) SyncMailbox(ctx context.Context, mailboxID string) error {
	if s.locker != nil {
		lockKey := "sync:" + mailboxID
		locked, err := s.locker.Lock(ctx, lockKey, 10*time.Minute)
		if err != nil {
			return fmt.Errorf("lock error: %w", err)
		}
		if !locked {
			return errors.New("mailbox already syncing")
		}
		defer s.locker.Unlock(ctx, lockKey)
	}

	mb, err := s.mailboxRepo.GetByID(ctx, mailboxID)
	if err != nil {
		return fmt.Errorf("get mailbox: %w", err)
	}
	if !mb.IsActive {
		return errors.New("mailbox is inactive")
	}

	if mb.TokenExpiry != nil && mb.TokenExpiry.Before(time.Now()) {
		if err := s.refreshMailboxToken(ctx, mb); err != nil {
			return fmt.Errorf("refresh token: %w", err)
		}
	}

	statusRunning, err := s.refRepo.GetSyncStatusID(ctx, "running")
	if err != nil {
		return fmt.Errorf("get running status: %w", err)
	}
	jobID := uuid.New().String()
	now := time.Now()
	job := &entity.SyncJob{
		ID:        jobID,
		MailboxID: mailboxID,
		StatusID:  statusRunning,
		StartedAt: &now,
		CreatedAt: now,
	}
	if err := s.syncJobRepo.Create(ctx, job); err != nil {
		return fmt.Errorf("create sync job: %w", err)
	}

	var messagesCount int
	var syncErr error
	func() {
		client := mail.NewIMAPClient()
		defer client.Close()

		if err := client.Connect(ctx, mb.Server, mb.Port, mb.UseTLS); err != nil {
			syncErr = fmt.Errorf("connect: %w", err)
			return
		}

		if err := client.Authenticate(ctx, mb); err != nil {
			syncErr = fmt.Errorf("auth: %w", err)
			return
		}

		lastUID, err := s.msgRepo.GetLastUID(ctx, mailboxID, "INBOX")
		if err != nil && !errors.Is(err, repoerrs.ErrNotFound) {
			syncErr = fmt.Errorf("get last uid: %w", err)
			return
		}

		messages, err := client.ListMessages(ctx, "INBOX", lastUID)
		if err != nil {
			syncErr = fmt.Errorf("list messages: %w", err)
			return
		}

		messagesCount = len(messages)
		for _, meta := range messages {
			full, err := client.FetchFullMessage(ctx, "INBOX", meta.UID)
			if err != nil {
				log.WithError(err).WithField("uid", meta.UID).Warn("fetch full message failed")
				continue
			}

			msg := &entity.Message{
				ID:             uuid.New().String(),
				MailboxID:      mailboxID,
				UID:            meta.UID,
				Folder:         "INBOX",
				Subject:        meta.Subject,
				From:           meta.From,
				To:             meta.To,
				Date:           parseDate(meta.Date),
				BodyPreview:    truncateString(full.BodyText, 200),
				HasAttachments: meta.HasAttachments,
				Seen:           meta.Seen,
				Flags:          meta.Flags,
				SyncedAt:       time.Now(),
				Hash:           computeHash(full.BodyText + full.BodyHTML),
			}
			if err := s.msgRepo.Upsert(ctx, msg); err != nil {
				log.WithError(err).WithField("uid", meta.UID).Warn("upsert message failed")
				continue
			}
			bodyEntity := &entity.MessageBody{
				MessageID: msg.ID,
				Body:      full.BodyText,
				BodyHTML:  full.BodyHTML,
			}
			if err := s.msgRepo.SaveBody(ctx, bodyEntity); err != nil {
				log.WithError(err).WithField("uid", meta.UID).Warn("save body failed")
			}
		}

		nowSync := time.Now()
		mb.LastSyncAt = &nowSync
		if err := s.mailboxRepo.Update(ctx, mb); err != nil {
			syncErr = fmt.Errorf("update mailbox: %w", err)
			return
		}
	}()

	statusSuccess, _ := s.refRepo.GetSyncStatusID(ctx, "success")
	statusFailed, _ := s.refRepo.GetSyncStatusID(ctx, "failed")
	finishTime := time.Now()
	job.FinishedAt = &finishTime
	job.MessagesCount = messagesCount
	if syncErr != nil {
		job.StatusID = statusFailed
		job.ErrorMsg = syncErr.Error()
	} else {
		job.StatusID = statusSuccess
	}
	if err := s.syncJobRepo.Update(ctx, job); err != nil {
		log.WithError(err).Error("update sync job failed")
	}

	return syncErr
}

func (s *SyncService) refreshMailboxToken(ctx context.Context, mb *entity.Mailbox) error {
	providerName := entity.IDToProvider[mb.ProviderID].String()
	tokenResp, err := s.oauthService.RefreshToken(ctx, providerName, mb.RefreshToken)
	if err != nil {
		return err
	}
	mb.AccessToken = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		mb.RefreshToken = tokenResp.RefreshToken
	}
	mb.TokenExpiry = ptrTime(time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second))
	mb.UpdatedAt = time.Now()
	return s.mailboxRepo.Update(ctx, mb)
}

func ptrTime(t time.Time) *time.Time { return &t }
func parseDate(dateStr string) time.Time {
	t, _ := time.Parse(time.RFC3339, dateStr)
	return t
}
func truncateString(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
func computeHash(text string) string {
	h := sha256.Sum256([]byte(text))
	return hex.EncodeToString(h[:])
}

func (s *SyncService) GetLastSyncStatus(ctx context.Context, mailboxID string) (*entity.SyncJob, error) {
	return s.syncJobRepo.GetLastByMailboxID(ctx, mailboxID)
}
