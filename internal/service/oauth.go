package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mail-sync-service/internal/config"
	"mail-sync-service/internal/entity"
	"mail-sync-service/internal/repo"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/mailru"
	"golang.org/x/oauth2/microsoft"
)

type OAuthService struct {
	config      *config.OAuth
	httpClient  *http.Client
	mailBoxRepo repo.Mailbox
}

func NewOAuthService(
	cfg *config.OAuth,
	httpClient *http.Client,
	mailBoxRepo repo.Mailbox,
) *OAuthService {
	return &OAuthService{
		config:      cfg,
		httpClient:  httpClient,
		mailBoxRepo: mailBoxRepo,
	}
}

type providerConfig struct {
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	RedirectURI  string
	Scopes       []string
}

// GetAuthURL возвращает URL для редиректа пользователя к провайдеру
func (s *OAuthService) GetAuthURL(ctx context.Context, provider string) (string, error) {
	cfg, err := s.getProviderConfig(provider)
	if err != nil {
		return "", err
	}
	oauth2Config := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURI,
		Scopes:       cfg.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.AuthURL,
			TokenURL: cfg.TokenURL,
		},
	}
	// Для разных провайдеров можно добавить параметры state, access_type и т.д.
	url := oauth2Config.AuthCodeURL("state", oauth2.AccessTypeOffline)
	return url, nil
}

// ExchangeCode обменивает код на токены
func (s *OAuthService) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	// Здесь нужно знать, от какого провайдера пришёл код.
	// Можно передавать provider в state или в отдельном параметре.
	// Для упрощения определяем по конфигурации (в реальном проекте лучше state).
	// Мы можем сохранить provider в сессии перед редиректом.
	// Пока вернём ошибку, если не знаем.
	return nil, errors.New("not implemented: need provider from state")
}

func (s *OAuthService) getProviderConfig(provider string) (*providerConfig, error) {
	switch provider {
	case "google":
		return &providerConfig{
			ClientID:     s.config.GoogleClientID,
			ClientSecret: s.config.GoogleClientSecret,
			AuthURL:      google.Endpoint.AuthURL,
			TokenURL:     google.Endpoint.TokenURL,
			RedirectURI:  s.config.RedirectURI,
			Scopes:       []string{"https://mail.google.com/"},
		}, nil
	case "microsoft":
		return &providerConfig{
			ClientID:     s.config.MicrosoftClientID,
			ClientSecret: s.config.MicrosoftClientSecret,
			AuthURL:      microsoft.AzureADEndpoint("common").AuthURL,
			TokenURL:     microsoft.AzureADEndpoint("common").TokenURL,
			RedirectURI:  s.config.RedirectURI,
			Scopes:       []string{"offline_access", "IMAP.AccessAsUser.All", "User.Read"},
		}, nil
	case "mailru":
		return &providerConfig{
			ClientID:     s.config.MailruClientID,
			ClientSecret: s.config.MailruClientSecret,
			AuthURL:      mailru.Endpoint.AuthURL,  // "https://o2.mail.ru/login"
			TokenURL:     mailru.Endpoint.TokenURL, // "https://o2.mail.ru/token"
			RedirectURI:  s.config.RedirectURI,
			Scopes:       []string{"mail.imap", "userinfo"}, // обязательные скоупы[reference:2]
		}, nil
	default:
		return nil, errors.New("unsupported provider: " + provider)
	}
}

// GetAuthURL генерирует URL для перенаправления пользователя
func (s *OAuthService) GetAuthURL(ctx context.Context, provider string) (string, error) {
	cfg, err := s.getProviderConfig(provider)
	if err != nil {
		return "", err
	}

	// Генерируем случайный state для CSRF-защиты (в реальном проекте нужно сохранить в сессии/БД)
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(stateBytes)

	// Формируем URL
	authURL, err := url.Parse(cfg.AuthURL)
	if err != nil {
		return "", err
	}
	q := authURL.Query()
	q.Set("client_id", cfg.ClientID)
	q.Set("redirect_uri", cfg.RedirectURI)
	q.Set("response_type", "code")
	q.Set("scope", strings.Join(cfg.Scopes, " "))
	q.Set("state", state)
	if provider == "google" {
		q.Set("access_type", "offline")
		q.Set("prompt", "consent")
	}
	authURL.RawQuery = q.Encode()
	return authURL.String(), nil
}

// HandleCallback обменивает код на токены и создаёт почтовый ящик
func (s *OAuthService) HandleCallback(ctx context.Context, provider, code, state string) (string, error) {
	cfg, err := s.getProviderConfig(provider)
	if err != nil {
		return "", err
	}

	// Обмен кода на токены
	tokenResp, err := s.exchangeCode(ctx, cfg, code)
	if err != nil {
		return "", err
	}

	// Извлекаем email пользователя (для Google/Microsoft можно запросить /userinfo)
	email, err := s.getUserEmail(ctx, provider, tokenResp.AccessToken)
	if err != nil {
		return "", err
	}

	// Создаём почтовый ящик
	mb := &entity.Mailbox{
		ID:           uuid.New().String(),
		Email:        email,
		Provider:     entity.Provider(provider),
		Protocol:     entity.ProtocolIMAP, // или определять по настройкам
		Server:       s.getServer(provider),
		Port:         s.getPort(provider),
		UseTLS:       true,
		AuthType:     entity.AuthTypeOAuth2,
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenExpiry:  ptrTime(time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)),
		IsActive:     true,
	}

	// Конвертируем строковые константы в ID
	providerID := entity.ProviderToID[mb.Provider]
	protocolID := entity.ProtocolToID[mb.Protocol]
	authTypeID := entity.AuthTypeToID[mb.AuthType]

	// Перекладываем в структуру с числовыми ID
	mbWithIDs := &entity.Mailbox{
		ID:           mb.ID,
		Email:        mb.Email,
		ProviderID:   providerID,
		ProtocolID:   protocolID,
		Server:       mb.Server,
		Port:         mb.Port,
		UseTLS:       mb.UseTLS,
		AuthTypeID:   authTypeID,
		AccessToken:  mb.AccessToken,
		RefreshToken: mb.RefreshToken,
		TokenExpiry:  mb.TokenExpiry,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	if err := s.mailboxRepo.Create(ctx, mbWithIDs); err != nil {
		return "", err
	}

	return mb.ID, nil
}

// exchangeCode обменивает код авторизации на токены
func (s *OAuthService) exchangeCode(ctx context.Context, cfg *providerConfig, code string) (*tokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", cfg.ClientID)
	data.Set("client_secret", cfg.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", cfg.RedirectURI)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, "POST", cfg.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed: %s", body)
	}

	var token tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}
	return &token, nil
}

// getUserEmail получает email пользователя (для Google, Microsoft)
func (s *OAuthService) getUserEmail(ctx context.Context, provider, accessToken string) (string, error) {
	var userInfoURL string
	switch provider {
	case "google":
		userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	case "microsoft":
		userInfoURL = "https://graph.microsoft.com/v1.0/me"
	default:
		return "", errors.New("unsupported provider for userinfo")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get user info: %d", resp.StatusCode)
	}

	var userInfo struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return "", err
	}
	if userInfo.Email == "" {
		return "", errors.New("email not found in userinfo response")
	}
	return userInfo.Email, nil
}

// вспомогательные функции
func (s *OAuthService) getServer(provider string) string {
	switch provider {
	case "google":
		return "imap.gmail.com"
	case "microsoft":
		return "outlook.office365.com"
	default:
		return ""
	}
}

func (s *OAuthService) getPort(provider string) int {
	return 993 // все используют SSL
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}
