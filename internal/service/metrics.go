package service

import (
	"context"
	"mail-sync-service/internal/entity"
	"mail-sync-service/internal/repo/clickhouse"
	"time"
)

type MetricsService struct {
	metricsRepo *clickhouse.MetricsRepo
}

func NewMetricsService(metricsRepo *clickhouse.MetricsRepo) *MetricsService {
	return &MetricsService{metricsRepo: metricsRepo}
}

func (s *MetricsService) SaveSyncMetric(ctx context.Context, metric *entity.SyncMetric) error {
	return s.metricsRepo.SaveSyncMetric(ctx, metric)
}

func (s *MetricsService) GetOverviewStats(ctx context.Context, since time.Time) (*clickhouse.AggregatedStats, error) {
	return s.metricsRepo.GetOverviewStats(ctx, since)
}
