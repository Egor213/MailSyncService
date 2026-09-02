package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mail-sync-service/internal/entity"
	"mail-sync-service/pkg/elastic"
)

type SearchRepo struct {
	client *elastic.Client
	index  string
}

func NewSearchRepo(client *elastic.Client, index string) *SearchRepo {
	return &SearchRepo{client: client, index: index}
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

type SearchResult struct {
	Total int64
	Hits  []struct {
		ID     string         `json:"_id"`
		Score  float64        `json:"_score"`
		Source entity.Message `json:"_source"`
	}
}

func (r *SearchRepo) Search(ctx context.Context, input SearchInput) (*SearchResult, error) {
	query := r.buildQuery(input)
	if input.Page < 1 {
		input.Page = 1
	}
	if input.Size < 1 {
		input.Size = 20
	}
	from := (input.Page - 1) * input.Size

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("encode query: %w", err)
	}

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex(r.index),
		r.client.Search.WithBody(&buf),
		r.client.Search.WithSize(input.Size),
		r.client.Search.WithFrom(from),
	)
	if err != nil {
		return nil, fmt.Errorf("elastic search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elastic error: %s", res.String())
	}

	var result struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				ID     string         `json:"_id"`
				Score  float64        `json:"_score"`
				Source entity.Message `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &SearchResult{
		Total: result.Hits.Total.Value,
		Hits:  result.Hits.Hits,
	}, nil
}

func (r *SearchRepo) IndexMessage(ctx context.Context, msg *entity.Message) error {
	doc := map[string]interface{}{
		"id":              msg.ID,
		"mailbox_id":      msg.MailboxID,
		"folder":          msg.Folder,
		"subject":         msg.Subject,
		"from_addr":       msg.From,
		"to_addr":         msg.To,
		"body":            msg.BodyPreview, // в реальности нужно передавать полный текст
		"body_html":       "",              // будет заполнено отдельно
		"date":            msg.Date,
		"has_attachments": msg.HasAttachments,
		"seen":            msg.Seen,
		"flags":           msg.Flags,
		"synced_at":       msg.SyncedAt,
	}
	_, err := r.client.Index(
		r.index,
		bytes.NewReader(mustJSON(doc)),
		r.client.Index.WithContext(ctx),
		r.client.Index.WithDocumentID(msg.ID),
	)
	return err
}

func (r *SearchRepo) IndexMessageBody(ctx context.Context, messageID, body, bodyHTML string) error {
	update := map[string]interface{}{
		"doc": map[string]interface{}{
			"body":      body,
			"body_html": bodyHTML,
		},
	}
	bodyBytes := mustJSON(update)
	_, err := r.client.Update(
		r.index,
		messageID,
		bytes.NewReader(bodyBytes),
		r.client.Update.WithContext(ctx),
	)
	return err
}

func (r *SearchRepo) GetBody(ctx context.Context, messageID string) (string, string, error) {
	res, err := r.client.Get(
		r.index,
		messageID,
		r.client.Get.WithContext(ctx),
	)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == 404 {
			return "", "", nil
		}
		return "", "", fmt.Errorf("elastic get error: %s", res.String())
	}
	var doc struct {
		Source struct {
			Body     string `json:"body"`
			BodyHTML string `json:"body_html"`
		} `json:"_source"`
	}
	if err := json.NewDecoder(res.Body).Decode(&doc); err != nil {
		return "", "", err
	}
	return doc.Source.Body, doc.Source.BodyHTML, nil
}

func (r *SearchRepo) buildQuery(input SearchInput) map[string]interface{} {
	must := []interface{}{}
	filter := []interface{}{}

	if input.Query != "" {
		must = append(must, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":     input.Query,
				"fields":    []string{"subject^3", "body", "from_addr^2", "to_addr^2"},
				"type":      "best_fields",
				"fuzziness": "AUTO",
			},
		})
	}
	if input.MailboxID != "" {
		filter = append(filter, map[string]interface{}{
			"term": map[string]interface{}{"mailbox_id": input.MailboxID},
		})
	}
	if input.Folder != "" {
		filter = append(filter, map[string]interface{}{
			"term": map[string]interface{}{"folder": input.Folder},
		})
	}
	if input.From != "" {
		filter = append(filter, map[string]interface{}{
			"match": map[string]interface{}{"from_addr": input.From},
		})
	}
	if input.To != "" {
		filter = append(filter, map[string]interface{}{
			"match": map[string]interface{}{"to_addr": input.To},
		})
	}
	if input.DateFrom != "" || input.DateTo != "" {
		rangeQuery := map[string]interface{}{}
		if input.DateFrom != "" {
			rangeQuery["gte"] = input.DateFrom
		}
		if input.DateTo != "" {
			rangeQuery["lte"] = input.DateTo
		}
		filter = append(filter, map[string]interface{}{
			"range": map[string]interface{}{"date": rangeQuery},
		})
	}
	if input.HasAttach != nil {
		filter = append(filter, map[string]interface{}{
			"term": map[string]interface{}{"has_attachments": *input.HasAttach},
		})
	}
	if input.Seen != nil {
		filter = append(filter, map[string]interface{}{
			"term": map[string]interface{}{"seen": *input.Seen},
		})
	}

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must":   must,
				"filter": filter,
			},
		},
	}
	return query
}

func mustJSON(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
