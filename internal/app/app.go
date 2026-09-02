package app

import (
	"context"
	"mail-sync-service/internal/config"
	httpapi "mail-sync-service/internal/controller/http/v1"
	"mail-sync-service/internal/repo"
	"mail-sync-service/internal/service"
	errutils "mail-sync-service/pkg/errors"
	"mail-sync-service/pkg/httpserver"
	"mail-sync-service/pkg/logger"
	"mail-sync-service/pkg/postgres"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/labstack/echo/v4"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

func Run() {
	// Config
	cfg, err := config.New()
	if err != nil {
		log.Fatal(errutils.WrapPathErr(err))
	}

	// Logger
	logger.SetupLogger(cfg.Log.Level)
	log.Info("Logger has been set up")

	// Migrations
	Migrate(cfg.PG.URL)

	// Redis
	// redisClient, err := redispkg.New(cfg.Redis.Address, cfg.Redis.Password, cfg.Redis.DB)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer redisClient.Close()

	// Kafka Producer
	// producer, err := kafka.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer producer.Close()

	// PostgreSQL – раскомментируем
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.MaxPoolSize))
	if err != nil {
		log.Fatal(err)
	}
	defer pg.Close()

	// Redis
	// redisClient, err := redispkg.New(cfg.Redis.Address, cfg.Redis.Password, cfg.Redis.DB)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer redisClient.Close()

	// Repos
	repositories := repo.NewRepositories(pg)

	// Transaction manager
	trManager := manager.Must(trmpgx.NewDefaultFactory(pg.Pool))

	// Redis
	redisClient, err := redispkg.New(cfg.Redis.Address, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatal(err)
	}
	defer redisClient.Close()
	locker := redis.NewRedisLocker(redisClient)

	// Kafka Producer
	producer, err := kafka.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic)
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()

	// Elasticsearch
	esClient, err := elastic.NewClient(cfg.ES.Address, cfg.ES.Username, cfg.ES.Password)
	if err != nil {
		log.Fatal(err)
	}
	searchRepo := elasticsearch.NewSearchRepo(esClient, cfg.ES.Index)

	// ClickHouse
	chClient, err := clickhouse.NewClient(cfg.ClickHouse.Address)
	if err != nil {
		log.Fatal(err)
	}
	metricsRepo := clickhouse.NewMetricsRepo(chClient)

	// Обновляем зависимости
	deps := service.ServicesDependencies{
		Repos:         repositories,
		Locker:        locker,
		KafkaProducer: producer,
		SearchRepo:    searchRepo,
		MetricsRepo:   metricsRepo,
		Config:        cfg,
		HTTPClient:    &http.Client{Timeout: 10 * time.Second},
		TrManager:     trManager,
	}
	services := service.NewServices(deps)

	// HTTP роутер (добавляем новые маршруты)
	e := echo.New()
	httpapi.ConfigureRouter(e, services)

	// Запуск Kafka Consumer (воркера)
	consumer, err := kafka.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.ConsumerGroup, cfg.Kafka.Topic)
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()

	// Создаём адаптер для syncService
	type syncHandlerAdapter struct {
		syncService *service.SyncService
	}
	func (a *syncHandlerAdapter) HandleSync(ctx context.Context, mailboxID string) error {
		return a.syncService.SyncMailbox(ctx, mailboxID)
	}
	handler := &syncHandlerAdapter{syncService: services.Sync.(*service.SyncService)}
	go func() {
		if err := consumer.Run(context.Background(), handler); err != nil {
			log.WithError(err).Error("kafka consumer stopped")
		}
	}()


	// Cron – периодическая синхронизация всех активных ящиков
	cronScheduler := cron.New(cron.WithChain(cron.Recover(cron.DefaultLogger)))
	_, err = cronScheduler.AddFunc("@every "+cfg.Sync.Interval.String(), func() {
		ctx := context.Background()
		log.Info("starting scheduled sync of all mailboxes")
		if err := services.Sync.SyncAllActive(ctx); err != nil {
			log.WithError(err).Error("scheduled sync failed")
		}
	})
	if err != nil {
		log.Fatal(err)
	}
	cronScheduler.Start()
	defer cronScheduler.Stop()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	select {
	case <-quit:
		log.Info("shutting down...")
	case err := <-httpServer.Notify():
		log.Error(err)
	}
	_ = httpServer.Shutdown()
}
