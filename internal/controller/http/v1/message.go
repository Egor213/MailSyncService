package httpapi

import (
	"errors"
	httpdto "mail-sync-service/internal/controller/http/v1/dto"
	repoerrs "mail-sync-service/internal/repo/errors"
	"mail-sync-service/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

type MessageHandler struct {
	searchService service.Search
}

func newMessageRoutes(g *echo.Group, searchService service.Search) {
	h := &MessageHandler{searchService: searchService}
	g.GET("/messages/:id/body", h.GetBody)
}

func (h *MessageHandler) GetBody(c echo.Context) error {
	id := c.Param("id")
	ctx := c.Request().Context()

	if _, err := h.searchService.GetMessageByID(ctx, id); err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return c.JSON(http.StatusNotFound, httpdto.ErrorOutput{Message: "message not found"})
		}
		return c.JSON(http.StatusInternalServerError, httpdto.ErrorOutput{Message: err.Error()})
	}

	body, bodyHTML, err := h.searchService.GetMessageBody(ctx, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, httpdto.ErrorOutput{Message: err.Error()})
	}
	if body == "" && bodyHTML == "" {
		return c.JSON(http.StatusNotFound, httpdto.ErrorOutput{Message: "message body not found"})
	}
	return c.JSON(http.StatusOK, map[string]string{
		"message_id": id,
		"body":       body,
		"body_html":  bodyHTML,
	})
}
