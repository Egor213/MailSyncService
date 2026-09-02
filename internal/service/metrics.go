package service

import (
	"context"
	"mail-sync-service/internal/entity"
	"mail-sync-service/internal/repo"
	"mail-sync-service/internal/repo/clickhouse"
	"time"
)

type MetricsService struct {
	metricsRepo repo.Metrics
}

func NewMetricsService(metricsRepo repo.Metrics) *MetricsService {
	return &MetricsService{metricsRepo: metricsRepo}
}

func (s *MetricsService) SaveSyncMetric(ctx context.Context, metric *entity.SyncMetric) error {
	return s.metricsRepo.SaveSyncMetric(ctx, metric)
}

func (s *MetricsService) GetOverviewStats(ctx context.Context, since time.Time) (*clickhouse.AggregatedStats, error) {
	return s.metricsRepo.GetOverviewStats(ctx, since)
}
