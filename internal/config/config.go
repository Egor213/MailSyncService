package config

import (
	"os"
	"time"

	errutils "mail-sync-service/pkg/errors"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	App        App        `yaml:"app"`
	HTTP       HTTP       `yaml:"http"`
	Log        Log        `yaml:"log"`
	PG         PG         `yaml:"postgres"`
	Redis      Redis      `yaml:"redis"`
	Kafka      Kafka      `yaml:"kafka"`
	OAuth      OAuth      `yaml:"oauth"`
	Sync       Sync       `yaml:"sync"`
	ES         ES         `yaml:"elasticsearch"`
	ClickHouse ClickHouse `yaml:"clickhouse"`
	Security   Security   `yaml:"security"`
}

type App struct {
	Name    string `yaml:"name" env-required:"true"`
	Version string `yaml:"version" env-required:"true"`
}

type HTTP struct {
	Address string `env-required:"true" env:"SERVER_ADDRESS"`
}

type Log struct {
	Level string `yaml:"level" env:"LOG_LEVEL" env-default:"info"`
}

type PG struct {
	URL         string `env-required:"true" env:"POSTGRES_CONN"`
	MaxPoolSize int    `env-required:"true" env:"MAX_POOL_SIZE" yaml:"max_pool_size"`
}

type Redis struct {
	Address  string `env-required:"true" env:"REDIS_ADDR"`
	Password string `env:"REDIS_PASSWORD"`
	DB       int    `env:"REDIS_DB" env-default:"0"`
}

type Kafka struct {
	Brokers       []string `yaml:"brokers" env:"KAFKA_BROKERS" env-separator:","`
	Topic         string   `yaml:"topic" env:"KAFKA_TOPIC" env-default:"mailbox-events"`
	ConsumerGroup string   `yaml:"consumer_group" env:"KAFKA_CONSUMER_GROUP" env-default:"mail-sync-group"`
}

type OAuth struct {
	GoogleClientID        string `env:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret    string `env:"GOOGLE_CLIENT_SECRET"`
	MicrosoftClientID     string `env:"MICROSOFT_CLIENT_ID"`
	MicrosoftClientSecret string `env:"MICROSOFT_CLIENT_SECRET"`
	MailruClientID        string `env:"MAILRU_CLIENT_ID"`
	MailruClientSecret    string `env:"MAILRU_CLIENT_SECRET"`
	YandexClientID        string `env:"YANDEX_CLIENT_ID"`
	YandexClientSecret    string `env:"YANDEX_CLIENT_SECRET"`
	RedirectURI           string `env:"OAUTH_REDIRECT_URI" env-default:"http://localhost:8080/api/v1/auth/callback"`
}

type Sync struct {
	Interval     time.Duration `yaml:"interval" env:"SYNC_INTERVAL" env-default:"5m"`
	RetryMax     int           `yaml:"retry_max" env:"SYNC_RETRY_MAX" env-default:"3"`
	RetryBackoff time.Duration `yaml:"retry_backoff" env:"SYNC_RETRY_BACKOFF" env-default:"2s"`
}

type Security struct {
	EncryptionKey string `env:"ENCRYPTION_KEY" env-required:"true"`
}

type ES struct {
	Addresses []string `yaml:"addresses" env:"ES_ADDRESSES" env-separator:","`
	Username  string   `yaml:"username" env:"ES_USERNAME"`
	Password  string   `yaml:"password" env:"ES_PASSWORD"`
	Index     string   `yaml:"index" env:"ES_INDEX" env-default:"mail_sync"`
}

type ClickHouse struct {
	Address string `yaml:"address" env:"CH_ADDRESS"`
}

func New() (*Config, error) {
	cfg := &Config{}
	if err := godotenv.Load("infra/.env"); err != nil {
		log.WithError(err).Info(".env file not found, using system environment")
	}
	pathToConfig, ok := os.LookupEnv("APP_CONFIG_PATH")
	if !ok || pathToConfig == "" {
		pathToConfig = "config/config.yaml"
	}
	if err := cleanenv.ReadConfig(pathToConfig, cfg); err != nil {
		return nil, errutils.WrapPathErr(err)
	}
	if err := cleanenv.UpdateEnv(cfg); err != nil {
		return nil, errutils.WrapPathErr(err)
	}
	return cfg, nil
}
