package httpapi

import (
	"encoding/base64"
	httpdto "mail-sync-service/internal/controller/http/v1/dto"
	"mail-sync-service/internal/service"
	"net/http"
	"strings"

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
	authURL, err := h.oauthService.GetAuthURL(c.Request().Context(), provider)
	if err != nil {
		return c.JSON(http.StatusBadRequest, httpdto.ErrorOutput{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"auth_url": authURL})
}

func (h *OAuthHandler) Callback(c echo.Context) error {
	code := c.QueryParam("code")
	state := c.QueryParam("state")
	if code == "" || state == "" {
		return c.JSON(http.StatusBadRequest, httpdto.ErrorOutput{Message: "missing code or state"})
	}

	stateBytes, err := base64.URLEncoding.DecodeString(state)
	if err != nil {
		return c.JSON(http.StatusBadRequest, httpdto.ErrorOutput{Message: "invalid state"})
	}
	stateStr := string(stateBytes)
	if !strings.HasPrefix(stateStr, "provider:") {
		return c.JSON(http.StatusBadRequest, httpdto.ErrorOutput{Message: "invalid state format"})
	}
	provider := strings.TrimPrefix(stateStr, "provider:")

	mailboxID, err := h.oauthService.HandleCallback(c.Request().Context(), provider, code, state)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, httpdto.ErrorOutput{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"mailbox_id": mailboxID})
}
