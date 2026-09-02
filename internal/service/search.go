package service

import (
	"context"
	"mail-sync-service/internal/entity"
	"mail-sync-service/internal/repo"
	"mail-sync-service/internal/repo/elasticsearch"
	"time"
)

type SearchService struct {
	searchRepo repo.Search
	msgRepo    repo.Message
}

func NewSearchService(searchRepo repo.Search, msgRepo repo.Message) *SearchService {
	return &SearchService{
		searchRepo: searchRepo,
		msgRepo:    msgRepo,
	}
}

type SearchInput struct {
	Query     string
	MailboxID string
	Folder    string
	From      string
	To        string
	DateFrom  string
	DateTo    string
	HasAttach *bool
	Seen      *bool
	Page      int
	Size      int
}

type SearchResultItem struct {
	ID             string    `json:"id"`
	MailboxID      string    `json:"mailbox_id"`
	Folder         string    `json:"folder"`
	Subject        string    `json:"subject"`
	From           string    `json:"from"`
	To             string    `json:"to"`
	Date           time.Time `json:"date"`
	BodyPreview    string    `json:"body_preview"`
	HasAttachments bool      `json:"has_attachments"`
	Seen           bool      `json:"seen"`
	Flags          []string  `json:"flags"`
	Score          float64   `json:"score"`
}

func (s *SearchService) Search(ctx context.Context, input SearchInput) ([]*SearchResultItem, int64, error) {
	esInput := elasticsearch.SearchInput{
		Query:     input.Query,
		MailboxID: input.MailboxID,
		Folder:    input.Folder,
		From:      input.From,
		To:        input.To,
		DateFrom:  input.DateFrom,
		DateTo:    input.DateTo,
		HasAttach: input.HasAttach,
		Seen:      input.Seen,
		Page:      input.Page,
		Size:      input.Size,
	}
	res, err := s.searchRepo.Search(ctx, esInput)
	if err != nil {
		return nil, 0, err
	}

	items := make([]*SearchResultItem, 0, len(res.Hits))
	for _, hit := range res.Hits {
		item := &SearchResultItem{
			ID:             hit.Source.ID,
			MailboxID:      hit.Source.MailboxID,
			Folder:         hit.Source.Folder,
			Subject:        hit.Source.Subject,
			From:           hit.Source.From,
			To:             hit.Source.To,
			Date:           hit.Source.Date,
			BodyPreview:    hit.Source.BodyPreview,
			HasAttachments: hit.Source.HasAttachments,
			Seen:           hit.Source.Seen,
			Flags:          hit.Source.Flags,
			Score:          hit.Score,
		}
		items = append(items, item)
	}
	return items, res.Total, nil
}

func (s *SearchService) GetMessageBody(ctx context.Context, messageID string) (string, string, error) {
	return s.searchRepo.GetBody(ctx, messageID)
}

func (s *SearchService) IndexMessage(ctx context.Context, msg *entity.Message) error {
	return s.searchRepo.IndexMessage(ctx, msg)
}

func (s *SearchService) IndexMessageBody(ctx context.Context, messageID, body, bodyHTML string) error {
	return s.searchRepo.IndexMessageBody(ctx, messageID, body, bodyHTML)
}
