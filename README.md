# Mail Sync Service

> Сервис фоновой синхронизации почтовых ящиков по IMAP: забирает новые письма, сохраняет метаданные и тела в PostgreSQL, индексирует письма в Elasticsearch, пишет метрики в ClickHouse и управляет синхронизацией через HTTP API, Kafka и cron.

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791?logo=postgresql)
![Redis](https://img.shields.io/badge/Redis-7+-DC382D?logo=redis)
![Kafka](https://img.shields.io/badge/Kafka-3+-231F20?logo=apachekafka)
![Elasticsearch](https://img.shields.io/badge/Elasticsearch-8+-005571?logo=elasticsearch)
![ClickHouse](https://img.shields.io/badge/ClickHouse-24+-FFCC01?logo=clickhouse)

---

## Содержание

- [О проекте](#о-проекте)
- [Возможности](#возможности)
- [Архитектура](#архитектура)
- [Технологии](#технологии)
- [Быстрый старт](#быстрый-старт)
- [Конфигурация](#конфигурация)
- [API](#api)
- [Как это работает](#как-это-работает)
- [Структура проекта](#структура-проекта)
- [Разработка](#разработка)
- [Безопасность](#безопасность)
- [Roadmap](#roadmap)
- [Лицензия](#лицензия)

---

## О проекте

**Mail Sync Service** — это backend-сервис на Go для синхронизации почтовых ящиков с поддержкой OAuth2 и обычной аутентификации. Сервис подключается к почтовым серверам по IMAP, забирает новые письма, сохраняет их в PostgreSQL, индексирует в Elasticsearch для полнотекстового поиска и агрегирует метрики синхронизации в ClickHouse.

Проект построен по принципам слоистой архитектуры:

- `controller/http` — HTTP-слой на Echo;
- `service` — бизнес-логика;
- `repo` — доступ к PostgreSQL, Elasticsearch, ClickHouse и Redis;
- `infrastruct` — внешние интеграции: Kafka, IMAP;
- `entity` — доменные модели и константы.

Сервис поддерживает:
- ручной запуск синхронизации через HTTP API;
- периодическую синхронизацию через cron;
- запуск синхронизации через Kafka-события;
- защиту от параллельной синхронизации одного ящика через Redis-лок;
- retry-логику при ошибках;
- graceful shutdown.

---

## Возможности

- **Синхронизация почты по IMAP**  
  Получение новых писем из папки `INBOX`, начиная с последнего известного UID.

- **OAuth2 для провайдеров**  
  Поддержка Google, Microsoft, Mail.ru, Yandex. Получение `access_token` и `refresh_token`, автоматическое обновление токена при истечении.

- **Хранение в PostgreSQL**  
  Почтовые ящики, письма, тела писем, задания синхронизации, справочники провайдеров/протоколов/типов авторизации.

- **Полнотекстовый поиск в Elasticsearch**  
  Индексация метаданных и тел писем. Поиск по теме, отправителю, получателю, телу, дате, вложениям и флагу прочтения.

- **Метрики в ClickHouse**  
  Сохранение метрик по каждой синхронизации: статус, количество писем, длительность, ошибка.

- **Kafka-события**  
  Producer публикует события `mailbox.sync`, consumer обрабатывает их и запускает синхронизацию.

- **Redis-локи**  
  Защита от одновременной синхронизации одного и того же ящика.

- **Cron-планировщик**  
  Периодический запуск синхронизации всех активных ящиков.

- **Шифрование токенов**  
  `access_token` и `refresh_token` шифруются перед сохранением в PostgreSQL.

- **Graceful shutdown**  
  Корректное завершение HTTP-сервера, Kafka consumer и cron.

---

## Архитектура

```mermaid
flowchart TB
    Client[HTTP Client] --> API[Echo API]
    Cron[Cron Scheduler] --> SyncSvc[Sync Service]
    KafkaConsumer[Kafka Consumer] --> SyncSvc
    API --> MailboxSvc[Mailbox Service]
    API --> OAuthSvc[OAuth Service]
    API --> SearchSvc[Search Service]
    API --> MetricsSvc[Metrics Service]

    MailboxSvc --> KafkaProducer[Kafka Producer]
    KafkaProducer --> Kafka[(Kafka)]

    SyncSvc --> IMAP[IMAP Server]
    SyncSvc --> PG[(PostgreSQL)]
    SyncSvc --> ES[(Elasticsearch)]
    SyncSvc --> CH[(ClickHouse)]
    SyncSvc --> Redis[(Redis Lock)]

    SearchSvc --> ES
    MetricsSvc --> CH
    OAuthSvc --> Providers[OAuth Providers]
```

### Поток синхронизации

1. Синхронизация запускается одним из способов:
   - HTTP-запрос `POST /api/v1/mailboxes/:id/sync`;
   - HTTP-запрос `POST /api/v1/sync/mailbox/:id`;
   - cron по расписанию `@every <SYNC_INTERVAL>`;
   - Kafka-событие `mailbox.sync`.

2. `SyncService.SyncMailbox` пытается захватить Redis-лок `sync:<mailbox_id>`.
3. Загружается почтовый ящик из PostgreSQL.
4. Если токен истёк — выполняется refresh через OAuth.
5. Создаётся запись `sync_jobs` со статусом `running`.
6. Устанавливается IMAP-соединение и выполняется аутентификация.
7. Запрашивается последний UID из PostgreSQL.
8. Загружаются новые письма, сохраняются в PostgreSQL.
9. Тела писем сохраняются в `message_bodies`.
10. Асинхронно выполняется индексация в Elasticsearch.
11. Обновляется `last_sync_at` у ящика.
12. Обновляется `sync_jobs` со статусом `success` или `failed`.
13. Асинхронно пишется метрика в ClickHouse.

---

## Технологии

| Слой | Технологии |
|---|---|
| Язык | Go |
| HTTP | Echo |
| Конфигурация | cleanenv, godotenv |
| PostgreSQL | pgx, Squirrel |
| Миграции | golang-migrate |
| Redis | go-redis |
| Kafka | IBM Sarama |
| Elasticsearch | go-elasticsearch |
| ClickHouse | clickhouse-go |
| Почта | go-imap/v2, go-sasl |
| OAuth | golang.org/x/oauth2 |
| Логирование | logrus |
| Планировщик | robfig/cron |
| Транзакции | avito-tech/go-transaction-manager |

---

## Быстрый старт

### Требования

- Go 1.21+
- PostgreSQL 15+
- Redis 7+
- Kafka 3+
- Elasticsearch 8+
- ClickHouse 24+
- Доступ к почтовому серверу по IMAP
- OAuth-приложения для выбранных провайдеров

### 1. Клонирование

```bash
git clone <repo-url>
cd mail-sync-service
```

### 2. Настройка окружения

Создайте файл `infra/.env`:

```env
APP_CONFIG_PATH=config/config.yaml
SERVER_ADDRESS=:8080
LOG_LEVEL=info

POSTGRES_CONN=postgres://user:password@localhost:5432/mail_sync?sslmode=disable
MAX_POOL_SIZE=10

REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=mailbox-events
KAFKA_CONSUMER_GROUP=mail-sync-group

GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
MICROSOFT_CLIENT_ID=
MICROSOFT_CLIENT_SECRET=
MAILRU_CLIENT_ID=
MAILRU_CLIENT_SECRET=
YANDEX_CLIENT_ID=
YANDEX_CLIENT_SECRET=
OAUTH_REDIRECT_URI=http://localhost:8080/api/v1/auth/callback

SYNC_INTERVAL=5m
SYNC_RETRY_MAX=3
SYNC_RETRY_BACKOFF=2s

ENCRYPTION_KEY=change-me-32-bytes-secret-key

ES_ADDRESSES=http://localhost:9200
ES_USERNAME=
ES_PASSWORD=
ES_INDEX=mail_sync

CH_ADDRESS=localhost:9000
```

Создайте `config/config.yaml`:

```yaml
app:
  name: mail-sync-service
  version: 0.1.0

http:
  address: ":8080"

log:
  level: info

postgres:
  max_pool_size: 10

redis:
  address: "localhost:6379"
  password: ""
  db: 0

kafka:
  brokers: ["localhost:9092"]
  topic: "mailbox-events"
  consumer_group: "mail-sync-group"

oauth:
  redirect_uri: "http://localhost:8080/api/v1/auth/callback"

sync:
  interval: "5m"
  retry_max: 3
  retry_backoff: "2s"

elasticsearch:
  addresses: ["http://localhost:9200"]
  username: ""
  password: ""
  index: "mail_sync"

clickhouse:
  address: "localhost:9000"
```

> `ENCRYPTION_KEY` обязателен. Используйте стойкий ключ и не коммитьте его в репозиторий.

### 3. Запуск зависимостей

Пример через Docker:

```bash
docker run -d --name mail-sync-postgres -p 5432:5432 \
  -e POSTGRES_USER=user \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=mail_sync \
  postgres:15

docker run -d --name mail-sync-redis -p 6379:6379 redis:7

docker run -d --name mail-sync-kafka -p 9092:9092 \
  -e KAFKA_CFG_NODE_ID=0 \
  -e KAFKA_CFG_PROCESS_ROLES=controller,broker \
  -e KAFKA_CFG_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093 \
  -e KAFKA_CFG_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
  -e KAFKA_CFG_CONTROLLER_QUORUM_VOTERS=0@localhost:9093 \
  bitnami/kafka:3

docker run -d --name mail-sync-es -p 9200:9200 \
  -e discovery.type=single-node \
  -e xpack.security.enabled=false \
  docker.elastic.co/elasticsearch/elasticsearch:8.13.0

docker run -d --name mail-sync-clickhouse -p 9000:9000 -p 8123:8123 clickhouse/clickhouse-server:24
```

### 4. Миграции

Миграции применяются автоматически при старте приложения в `app.Run()` через `golang-migrate`.

Вручную:

```bash
migrate -path migrations -database "postgres://user:password@localhost:5432/mail_sync?sslmode=disable" up
```

### 5. Запуск

```bash
go mod download
go run ./cmd/app
```

Если entrypoint находится в другом месте, используйте свой путь, например:

```bash
go run .
```

После запуска:

- HTTP API: `http://localhost:8080`
- Healthcheck: `GET /health`

---

## Конфигурация

### Переменные окружения

| Переменная | Обязательна | По умолчанию | Описание |
|---|---:|---:|---|
| `APP_CONFIG_PATH` | нет | `config/config.yaml` | Путь к YAML-конфигу |
| `SERVER_ADDRESS` | да | — | Адрес HTTP-сервера |
| `LOG_LEVEL` | нет | `info` | Уровень логирования |
| `POSTGRES_CONN` | да | — | DSN для PostgreSQL |
| `MAX_POOL_SIZE` | да | — | Размер пула подключений PostgreSQL |
| `REDIS_ADDR` | да | — | Адрес Redis |
| `REDIS_PASSWORD` | нет | — | Пароль Redis |
| `REDIS_DB` | нет | `0` | Номер БД Redis |
| `KAFKA_BROKERS` | нет | — | Список брокеров Kafka через запятую |
| `KAFKA_TOPIC` | нет | `mailbox-events` | Топик Kafka |
| `KAFKA_CONSUMER_GROUP` | нет | `mail-sync-group` | Consumer group |
| `GOOGLE_CLIENT_ID` | нет | — | Google OAuth Client ID |
| `GOOGLE_CLIENT_SECRET` | нет | — | Google OAuth Client Secret |
| `MICROSOFT_CLIENT_ID` | нет | — | Microsoft OAuth Client ID |
| `MICROSOFT_CLIENT_SECRET` | нет | — | Microsoft OAuth Client Secret |
| `MAILRU_CLIENT_ID` | нет | — | Mail.ru OAuth Client ID |
| `MAILRU_CLIENT_SECRET` | нет | — | Mail.ru OAuth Client Secret |
| `YANDEX_CLIENT_ID` | нет | — | Yandex OAuth Client ID |
| `YANDEX_CLIENT_SECRET` | нет | — | Yandex OAuth Client Secret |
| `OAUTH_REDIRECT_URI` | нет | `http://localhost:8080/api/v1/auth/callback` | Redirect URI |
| `SYNC_INTERVAL` | нет | `5m` | Интервал cron-синхронизации |
| `SYNC_RETRY_MAX` | нет | `3` | Максимум retry |
| `SYNC_RETRY_BACKOFF` | нет | `2s` | Базовая задержка retry |
| `ENCRYPTION_KEY` | да | — | Ключ шифрования токенов |
| `ES_ADDRESSES` | нет | — | Адреса Elasticsearch через запятую |
| `ES_USERNAME` | нет | — | Логин Elasticsearch |
| `ES_PASSWORD` | нет | — | Пароль Elasticsearch |
| `ES_INDEX` | нет | `mail_sync` | Индекс Elasticsearch |
| `CH_ADDRESS` | нет | — | Адрес ClickHouse |

---

## API

Базовый префикс: `/api/v1`

### Health

| Метод | Путь | Описание |
|---|---|---|
| `GET` | `/health` | Проверка живости |

### OAuth

| Метод | Путь | Описание |
|---|---|---|
| `GET` | `/auth/login/:provider` | Получить URL для OAuth-авторизации |
| `GET` | `/auth/callback` | Callback OAuth |

Поддерживаемые `provider`: `google`, `microsoft`, `mailru`, `yandex`.

Пример:

```bash
curl http://localhost:8080/api/v1/auth/login/google
```

Ответ:

```json
{
  "auth_url": "https://accounts.google.com/o/oauth2/auth?..."
}
```

Callback:

```bash
curl "http://localhost:8080/api/v1/auth/callback?code=...&state=..."
```

Ответ:

```json
{
  "mailbox_id": "uuid"
}
```

### Почтовые ящики

| Метод | Путь | Описание |
|---|---|---|
| `POST` | `/mailboxes` | Создать ящик |
| `GET` | `/mailboxes` | Список активных ящиков |
| `GET` | `/mailboxes/:id` | Получить ящик |
| `PUT` | `/mailboxes/:id` | Обновить ящик |
| `DELETE` | `/mailboxes/:id` | Удалить ящик |
| `POST` | `/mailboxes/:id/sync` | Запустить синхронизацию |

Пример создания:

```bash
curl -X POST http://localhost:8080/api/v1/mailboxes \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "user@example.com",
    "provider": "gmail",
    "protocol": "imap",
    "server": "imap.gmail.com",
    "port": 993,
    "use_tls": true,
    "auth_type": "oauth2",
    "access_token": "ya29...",
    "refresh_token": "1//..."
  }'
```

### Синхронизация

| Метод | Путь | Описание |
|---|---|---|
| `POST` | `/sync/mailbox/:id` | Запустить синхронизацию |
| `GET` | `/sync/mailbox/:id/status` | Последний статус синхронизации |

Пример:

```bash
curl -X POST http://localhost:8080/api/v1/sync/mailbox/<mailbox_id>
curl http://localhost:8080/api/v1/sync/mailbox/<mailbox_id>/status
```

> В текущей реализации HTTP-запуск синхронизации выполняется синхронно в рамках запроса. Для асинхронного запуска используйте Kafka-события или cron.

### Поиск

| Метод | Путь | Описание |
|---|---|---|
| `GET` | `/search` | Поиск писем |

Query-параметры:

| Параметр | Описание |
|---|---|
| `q` | Поисковый запрос |
| `mailbox_id` | ID ящика |
| `folder` | Папка |
| `from` | Отправитель |
| `to` | Получатель |
| `date_from` | Дата от |
| `date_to` | Дата до |
| `has_attachments` | Есть вложения |
| `seen` | Прочитано |
| `page` | Страница |
| `size` | Размер страницы |

Пример:

```bash
curl 'http://localhost:8080/api/v1/search?q=invoice&mailbox_id=<id>&page=1&size=20'
```

### Сообщения

| Метод | Путь | Описание |
|---|---|---|
| `GET` | `/messages/:id/body` | Получить тело письма |

Пример:

```bash
curl http://localhost:8080/api/v1/messages/<message_id>/body
```

### Метрики

| Метод | Путь | Описание |
|---|---|---|
| `GET` | `/metrics/overview` | Агрегированные метрики |

Query-параметры:

| Параметр | Описание |
|---|---|
| `since` | Начало периода в формате RFC3339 |

Пример:

```bash
curl 'http://localhost:8080/api/v1/metrics/overview?since=2024-01-01T00:00:00Z'
```

---

## Как это работает

### Синхронизация

Основная логика находится в `internal/service/sync.go`.

Алгоритм:

1. Захват Redis-лока `sync:<mailbox_id>` на 10 минут.
2. Загрузка ящика из PostgreSQL.
3. Проверка `is_active`.
4. Проверка срока действия `token_expiry`.
5. При необходимости — refresh OAuth-токена.
6. Создание `sync_jobs` со статусом `running`.
7. Подключение к IMAP.
8. Аутентификация:
   - `XOAUTH2` для OAuth2;
   - `PLAIN` для логина/пароля.
9. Получение последнего UID из `messages`.
10. Загрузка новых писем батчами по 50.
11. Сохранение метаданных в `messages`.
12. Сохранение тела в `message_bodies`.
13. Асинхронная индексация в Elasticsearch.
14. Обновление `last_sync_at`.
15. Обновление `sync_jobs`.
16. Асинхронная запись метрики в ClickHouse.

Retry:

- до `SYNC_RETRY_MAX` попыток;
- задержка `SYNC_RETRY_BACKOFF * attempt`;
- ошибки `ErrMailboxInactive` и `ErrAlreadySyncing` не ретраятся.

### OAuth

Логика в `internal/service/oauth.go`.

1. `GET /auth/login/:provider` формирует URL авторизации.
2. Пользователь переходит по URL и подтверждает доступ.
3. Провайдер редиректит на `/auth/callback?code=...&state=...`.
4. Сервис обменивает `code` на `access_token` и `refresh_token`.
5. Получает email пользователя через userinfo endpoint.
6. Создаёт почтовый ящик с `auth_type=oauth2`.
7. Токены шифруются и сохраняются в PostgreSQL.

### Kafka

Producer:

- публикует событие `mailbox.sync` с `mailbox_id`;
- используется для асинхронного запуска синхронизации.

Consumer:

- читает топик `mailbox-events`;
- вызывает `HandleSync`;
- после обработки делает `MarkMessage`.

### Поиск

Elasticsearch-репозиторий:

- индексирует метаданные письма;
- отдельно обновляет тело и HTML-тело;
- поддерживает фильтры по ящику, папке, отправителю, получателю, дате, вложениям и прочтению;
- использует `multi_match` с boost по полям.

### Метрики

ClickHouse-репозиторий:

- пишет `sync_metrics`;
- агрегирует статистику по провайдерам;
- считает общее количество синхронизаций, успешных, неуспешных, среднюю длительность и количество писем.

---

## Структура проекта

```text
.
├── cmd/                          # entrypoint приложения
├── config/                       # config.yaml
├── infra/                        # .env и инфраструктурные файлы
├── internal/
│   ├── app/                      # запуск, миграции, graceful shutdown
│   ├── config/                   # загрузка конфигурации
│   ├── controller/
│   │   └── http/
│   │       └── v1/               # Echo handlers, DTO, middleware
│   ├── entity/                   # доменные модели и константы
│   ├── infrastruct/
│   │   ├── kafka/                # producer и consumer
│   │   └── mail/                 # IMAP-клиент
│   ├── repo/
│   │   ├── clickhouse/           # метрики
│   │   ├── elasticsearch/        # поиск
│   │   ├── errors/               # ошибки репозиториев
│   │   ├── pgdb/                 # PostgreSQL
│   │   └── redis/                # локи
│   └── service/                  # бизнес-логика
├── migrations/                   # SQL-миграции
├── pkg/                          # переиспользуемые пакеты
│   ├── clickhouse/
│   ├── crypto/
│   ├── elastic/
│   ├── errors/
│   ├── httpserver/
│   ├── logger/
│   ├── postgres/
│   └── redis/
└── README.md
```

---

## Разработка

### Сборка

```bash
go build ./...
```

### Тесты

```bash
go test ./...
```

### Статический анализ

```bash
go vet ./...
```

### Форматирование

```bash
gofmt -w .
```

### Миграции

Применить:

```bash
migrate -path migrations -database "$POSTGRES_CONN?sslmode=disable" up
```

Откатить:

```bash
migrate -path migrations -database "$POSTGRES_CONN?sslmode=disable" down
```

### Логи

HTTP-логи пишутся в stdout и в `logs/logfile.log`.

---

## Безопасность

- `access_token` и `refresh_token` шифруются с использованием `ENCRYPTION_KEY`.
- Не коммитьте `infra/.env`, `config/config.yaml` и реальные секреты.
- Используйте разные ключи для dev/stage/prod.
- Ограничьте доступ к PostgreSQL, Redis, Kafka, Elasticsearch и ClickHouse.
- Для production включите TLS для PostgreSQL и Elasticsearch.
- OAuth `state` содержит только `provider:<name>`. Для production рекомендуется добавить CSRF-защиту и подпись state.

---

## Roadmap

- [ ] OpenAPI/Swagger-документация.
- [ ] Аутентификация и авторизация HTTP API.
- [ ] Поддержка POP3.
- [ ] Синхронизация папок кроме `INBOX`.
- [ ] Загрузка и хранение вложений.
- [ ] Prometheus-метрики и Grafana-дашборды.
- [ ] Unit- и integration-тесты.
- [ ] CI/CD pipeline.
- [ ] Идемпотентная обработка Kafka-сообщений.
- [ ] Улучшенный MIME-парсер.
- [ ] Контекстные таймауты для внешних вызовов.

---