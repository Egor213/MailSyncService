package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	// Замените этот импорт на реальный путь к пакету mail в вашем проекте
	"mail-sync-service/internal/entity"
	"mail-sync-service/internal/infrastruct/mail"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2/yandex"
)

func main() {
	// ==========================================================
	// НАСТРОЙКИ ДЛЯ ТЕСТА
	// Заполните эти значения или передайте через переменные окружения
	// ==========================================================
	server := getEnv("IMAP_SERVER", "imap.yandex.com")
	port := 993
	useTLS := true
	email := "egor.muravkini@yandex.ru" // Ваш email, например "test@yandex.ru"
	token := ""                         // Refresh токен (y0__...) или Access токен (y0_Ag...) или Пароль приложения (для Plain)

	// if token != "" && isYandexRefresh(token) {
	// 	fmt.Println("0. Обнаружен refresh токен, получаю access токен...")
	// 	access, err := refreshAccessToken(token)
	// 	if err != nil {
	// 		log.Fatalf("❌ Ошибка обмена refresh токена: %v", err)
	// 	}
	// 	token = access
	// 	fmt.Println("✅ Access токен получен")
	// }

	// Выберите тип аутентификации:
	// 1. Для OAuth2 (как в вашем основном коде):
	// authTypeID := entity.AuthTypeToID[entity.AuthTypeOAuth2]
	authTypeID := entity.AuthTypeToID[entity.AuthTypePlain]

	// 2. Для проверки паролем приложения (Plain), раскомментируйте строку ниже:
	// authTypeID := entity.AuthTypeToID[entity.AuthTypePlain]
	// ==========================================================

	if email == "" || token == "" {
		log.Fatal("Ошибка: Пожалуйста, установите переменные окружения IMAP_EMAIL и IMAP_TOKEN или задайте их в коде напрямую.")
	}

	ctx := context.Background()
	client := mail.NewIMAPClient()

	fmt.Println("1. Подключение к серверу...")
	err := client.Connect(ctx, server, port, useTLS)
	if err != nil {
		log.Fatalf("❌ Ошибка подключения: %v", err)
	}
	fmt.Println("✅ Успешно подключено!")
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Предупреждение при закрытии соединения: %v", err)
		}
	}()

	fmt.Println("2. Аутентификация...")
	mailbox := &entity.Mailbox{
		Email:       email,
		AccessToken: token, // В случае Plain auth это поле работает как пароль
		AuthTypeID:  authTypeID,
	}

	err = client.Authenticate(ctx, mailbox)
	if err != nil {
		log.Fatalf("❌ Ошибка аутентификации: %v", err)
	}
	fmt.Println("✅ Успешно аутентифицировано!")

	fmt.Println("3. Получение списка сообщений из INBOX...")
	// Передаем пустой lastUID, чтобы получить доступные сообщения
	messages, err := client.ListMessages(ctx, "INBOX", "")
	if err != nil {
		log.Fatalf("❌ Ошибка получения списка сообщений: %v", err)
	}

	if len(messages) == 0 {
		fmt.Println("⚠️ В ящике нет сообщений.")
		return
	}

	fmt.Printf("✅ Найдено сообщений: %d\n", len(messages))

	// Берем последнее сообщение (обычно они приходят в порядке возрастания UID, так что последнее в срезе - самое свежее)
	lastMsg := messages[len(messages)-1]

	fmt.Println("\n--- Метаданные последнего сообщения ---")
	fmt.Printf("UID: %s\n", lastMsg.UID)
	fmt.Printf("From: %s\n", lastMsg.From)
	fmt.Printf("To: %s\n", lastMsg.To)
	fmt.Printf("Subject: %s\n", lastMsg.Subject)
	fmt.Printf("Date: %s\n", lastMsg.Date)
	fmt.Printf("Seen: %v\n", lastMsg.Seen)

	fmt.Println("\n4. Получение полного текста последнего сообщения...")
	fullMsg, err := client.FetchFullMessage(ctx, "INBOX", lastMsg.UID)
	if err != nil {
		log.Fatalf("❌ Ошибка получения полного сообщения: %v", err)
	}

	fmt.Println("--- Тело сообщения (первые 500 символов) ---")
	if fullMsg.BodyText != "" {
		limit := 500
		if len(fullMsg.BodyText) < limit {
			limit = len(fullMsg.BodyText)
		}
		fmt.Printf("[TEXT]:\n%s\n", fullMsg.BodyText[:limit])
	} else if fullMsg.BodyHTML != "" {
		limit := 500
		if len(fullMsg.BodyHTML) < limit {
			limit = len(fullMsg.BodyHTML)
		}
		fmt.Printf("[HTML]:\n%s\n", fullMsg.BodyHTML[:limit])
	} else {
		fmt.Println("⚠️ Тело сообщения пустое или не удалось извлечь текст/HTML.")
	}

	fmt.Println("\n🎉 Тест IMAP завершен успешно!")
}

// Вспомогательная функция для получения переменных окружения с значением по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// Yandex refresh токен начинается с "y0__", access токен — с "y0_Ag"
func isYandexRefresh(token string) bool {
	return len(token) > 4 && token[:4] == "y0__"
}

func refreshAccessToken(refreshToken string) (string, error) {
	_ = godotenv.Load("infra/.env")
	clientID := os.Getenv("YANDEX_CLIENT_ID")
	clientSecret := os.Getenv("YANDEX_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("YANDEX_CLIENT_ID/SECRET не заданы в infra/.env")
	}

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")

	resp, err := http.PostForm(yandex.Endpoint.TokenURL, data)
	if err != nil {
		return "", fmt.Errorf("запрос не удался: %w", err)
	}
	defer resp.Body.Close()

	var body struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		Description string `json:"error_description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("не удалось распарсить ответ: %w", err)
	}
	if body.AccessToken == "" {
		return "", fmt.Errorf("обмен не удался: %s %s (status %d)", body.Error, body.Description, resp.StatusCode)
	}
	return body.AccessToken, nil
}
