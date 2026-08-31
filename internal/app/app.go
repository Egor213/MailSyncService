package app

import (
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

	// Repos
	repositories := repo.NewRepositories(pg)

	// Transaction manager
	trManager := manager.Must(trmpgx.NewDefaultFactory(pg.Pool))

	// Services
	deps := service.ServicesDependencies{
		Repos:      repositories,
		TrManager:  trManager,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		Config:     cfg,
	}
	services := service.NewServices(deps)

	// HTTP
	e := echo.New()
	httpapi.ConfigureRouter(e, services)

	httpServer := httpserver.New(e, httpserver.Address(cfg.HTTP.Address))

	// Kafka Consumer (воркер)
	// consumer, err := kafka.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.ConsumerGroup, cfg.Kafka.Topic)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer consumer.Close()
	// syncWorker := worker.NewSyncWorker(consumer, syncService, cfg.Kafka.Topic)
	// go syncWorker.Run(context.Background())

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
