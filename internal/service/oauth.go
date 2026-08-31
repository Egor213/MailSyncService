package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mail-sync-service/internal/config"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/mailru"
	"golang.org/x/oauth2/microsoft"
	"golang.org/x/oauth2/yandex"
)

type OAuthService struct {
	config         *config.OAuth
	httpClient     *http.Client
	mailboxService Mailbox
}

func NewOAuthService(
	cfg *config.OAuth,
	httpClient *http.Client,
	mailboxService Mailbox,
) *OAuthService {
	return &OAuthService{
		config:         cfg,
		httpClient:     httpClient,
		mailboxService: mailboxService,
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
			AuthURL:      mailru.Endpoint.AuthURL,
			TokenURL:     mailru.Endpoint.TokenURL,
			RedirectURI:  s.config.RedirectURI,
			Scopes:       []string{"mail.imap", "userinfo"},
		}, nil
	case "yandex":
		return &providerConfig{
			ClientID:     s.config.YandexClientID,
			ClientSecret: s.config.YandexClientSecret,
			AuthURL:      yandex.Endpoint.AuthURL,
			TokenURL:     yandex.Endpoint.TokenURL,
			RedirectURI:  s.config.RedirectURI,
			Scopes:       []string{},
		}, nil
	default:
		return nil, errors.New("unsupported provider: " + provider)
	}
}

func (s *OAuthService) GetAuthURL(ctx context.Context, provider string) (string, error) {
	cfg, err := s.getProviderConfig(provider)
	if err != nil {
		return "", err
	}

	stateData := "provider:" + provider
	state := base64.URLEncoding.EncodeToString([]byte(stateData))

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

func (s *OAuthService) HandleCallback(ctx context.Context, provider, code, state string) (string, error) {
	cfg, err := s.getProviderConfig(provider)
	if err != nil {
		return "", err
	}

	tokenResp, err := s.exchangeCode(ctx, cfg, code)
	if err != nil {
		return "", err
	}

	email, err := s.getUserEmail(ctx, provider, tokenResp.AccessToken)
	if err != nil {
		return "", err
	}

	now := time.Now()
	expiry := now.Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	input := CreateMailboxInput{
		Email:        email,
		Provider:     provider,
		Protocol:     "imap",
		Server:       s.getServer(provider),
		Port:         s.getPort(provider),
		UseTLS:       true,
		AuthType:     "oauth2",
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenExpiry:  &expiry,
	}

	mailbox, err := s.mailboxService.CreateMailbox(ctx, input)
	if err != nil {
		return "", err
	}
	return mailbox.ID, nil
}

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
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
			return nil, fmt.Errorf("token exchange failed: %s - %s", errResp.Error, errResp.ErrorDescription)
		}
		return nil, fmt.Errorf("token exchange failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var token tokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}
	return &token, nil
}

func (s *OAuthService) getUserEmail(ctx context.Context, provider, accessToken string) (string, error) {
	var userInfoURL string
	switch provider {
	case "google":
		userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	case "microsoft":
		userInfoURL = "https://graph.microsoft.com/v1.0/me"
	case "yandex":
		userInfoURL = "https://login.yandex.ru/info?format=json"
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
		Email        string `json:"email"`
		DefaultEmail string `json:"default_email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return "", err
	}

	email := userInfo.Email
	if email == "" && provider == "yandex" {
		email = userInfo.DefaultEmail
	}
	if email == "" {
		return "", errors.New("email not found in userinfo response")
	}
	return email, nil

}

func (s *OAuthService) getServer(provider string) string {
	switch provider {
	case "google":
		return "imap.gmail.com"
	case "microsoft":
		return "outlook.office365.com"
	case "yandex":
		return "imap.yandex.ru"
	default:
		return ""
	}
}

func (s *OAuthService) getPort(provider string) int {
	return 993
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}
