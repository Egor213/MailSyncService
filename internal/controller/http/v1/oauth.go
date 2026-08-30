package httpapi

import (
	httpdto "mail-sync-service/internal/controller/http/v1/dto"
	"mail-sync-service/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

type OAuthHandler struct {
	oauthService service.OAuth
}

func newOAuthRoutes(g *echo.Group, oauthService service.OAuth) {
	h := &OAuthHandler{oauthService: oauthService}

	g.GET("/login/:provider", h.Login)
	g.GET("/callback", h.Callback)
}

func (h *OAuthHandler) Login(c echo.Context) error {
	provider := c.Param("provider")
	authURL, err := h.oauthService.GetAuthURL(provider)
	if err != nil {
		return c.JSON(http.StatusBadRequest, httpdto.ErrorOutput{Message: err.Error()})
	}
	return c.Redirect(http.StatusFound, authURL)
}

// Callback обрабатывает ответ от провайдера с кодом авторизации
func (h *OAuthHandler) Callback(c echo.Context) error {
	code := c.QueryParam("code")
	if code == "" {
		return c.JSON(http.StatusBadRequest, httpdto.ErrorOutput{Message: "missing code"})
	}
	// В реальном приложении нужно определить провайдера из state или из запроса,
	// здесь упрощённо – сохраняем полученные токены и возвращаем их пользователю.
	tokens, err := h.oauthService.ExchangeCode(c.Request().Context(), code)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, httpdto.ErrorOutput{Message: err.Error()})
	}
	// Возвращаем токены клиенту (можно отдать в JSON или сохранить сессию)
	return c.JSON(http.StatusOK, dto.OAuthTokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	})
}
