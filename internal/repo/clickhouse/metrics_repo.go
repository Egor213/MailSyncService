package clickhouse

import (
	"context"
	"mail-sync-service/internal/entity"
	"mail-sync-service/pkg/clickhouse"
	"time"
)

type MetricsRepo struct {
	client *clickhouse.Client
}

func NewMetricsRepo(client *clickhouse.Client) *MetricsRepo {
	return &MetricsRepo{client: client}
}

func (r *MetricsRepo) SaveSyncMetric(ctx context.Context, metric *entity.SyncMetric) error {
	query := `
		INSERT INTO mail_sync.sync_metrics 
		(mailbox_id, provider, protocol, status, messages_count, duration_ms, error_msg, started_at, finished_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.client.Exec(ctx, query,
		metric.MailboxID,
		metric.Provider,
		metric.Protocol,
		metric.Status,
		metric.MessagesCount,
		metric.DurationMs,
		metric.ErrorMsg,
		metric.StartedAt,
		metric.FinishedAt,
	)
	return err
}

type AggregatedStats struct {
	TotalSyncs    int64            `json:"total_syncs"`
	SuccessSyncs  int64            `json:"success_syncs"`
	FailedSyncs   int64            `json:"failed_syncs"`
	AvgDurationMs float64          `json:"avg_duration_ms"`
	TotalMessages int64            `json:"total_messages"`
	ByProvider    map[string]int64 `json:"by_provider"`
}

func (r *MetricsRepo) GetOverviewStats(ctx context.Context, since time.Time) (*AggregatedStats, error) {
	query := `
		SELECT 
			count() AS total,
			sumIf(1, status='success') AS success,
			sumIf(1, status='failed') AS failed,
			avg(duration_ms) AS avg_duration,
			sum(messages_count) AS total_msgs,
			provider
		FROM mail_sync.sync_metrics
		WHERE created_at >= ?
		GROUP BY provider
	`
	rows, err := r.client.Query(ctx, query, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := &AggregatedStats{
		ByProvider: make(map[string]int64),
	}
	for rows.Next() {
		var provider string
		var total, success, failed, totalMsgs int64
		var avgDuration float64
		if err := rows.Scan(&total, &success, &failed, &avgDuration, &totalMsgs, &provider); err != nil {
			return nil, err
		}
		stats.TotalSyncs += total
		stats.SuccessSyncs += success
		stats.FailedSyncs += failed
		stats.AvgDurationMs += avgDuration * float64(total) // weighted average позже
		stats.TotalMessages += totalMsgs
		stats.ByProvider[provider] = total
	}
	if stats.TotalSyncs > 0 {
		stats.AvgDurationMs /= float64(stats.TotalSyncs)
	}
	return stats, nil
}
