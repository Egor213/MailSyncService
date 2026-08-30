package app

import (
	"mail-sync-service/internal/config"
	httpapi "mail-sync-service/internal/controller/http/v1"
	"mail-sync-service/internal/service"
	errutils "mail-sync-service/pkg/errors"
	"mail-sync-service/pkg/httpserver"
	"mail-sync-service/pkg/logger"
	"os"
	"os/signal"
	"syscall"

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

	// PostgreSQL
	// pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.MaxPoolSize))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer pg.Close()

	// Redis
	// redisClient, err := redispkg.New(cfg.Redis.Address, cfg.Redis.Password, cfg.Redis.DB)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer redisClient.Close()

	// Репозитории
	// mailboxRepo := pgdb.NewMailboxRepo(pg)
	// msgRepo := pgdb.NewMessageRepo(pg)
	// syncJobRepo := pgdb.NewSyncJobRepo(pg)
	// locker := redis.NewLocker(redisClient)

	// Kafka Producer
	// producer, err := kafka.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer producer.Close()

	// Сервисы
	// mailboxService := service.NewMailboxService(mailboxRepo, msgRepo, syncJobRepo, locker, producer)
	// syncService := service.NewSyncService(mailboxRepo, msgRepo, syncJobRepo, locker, nil, nil) // с фабриками

	// HTTP
	e := echo.New()
	httpapi.ConfigureRouter(e, &service.Services{Mailbox: mailboxService})

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
