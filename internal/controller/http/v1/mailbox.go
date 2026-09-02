package httpapi

import (
	"errors"
	httpdto "mail-sync-service/internal/controller/http/v1/dto"
	repoerrs "mail-sync-service/internal/repo/errors"
	"mail-sync-service/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

type MailboxHandler struct {
	mailboxService service.Mailbox
	syncService    service.Sync
}

func newMailboxRoutes(g *echo.Group, mailboxService service.Mailbox, syncService service.Sync) {
	h := &MailboxHandler{mailboxService: mailboxService, syncService: syncService}

	g.POST("", h.Create)               // POST /api/v1/mailboxes
	g.POST("/:id/sync", h.TriggerSync) // POST /api/v1/mailboxes/:id/sync
}

func (h *MailboxHandler) Create(c echo.Context) error {
	var req httpdto.CreateMailboxRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httpdto.ErrorOutput{Message: "invalid request"})
	}

	mb, err := h.mailboxService.CreateMailbox(c.Request().Context(), service.CreateMailboxInput{
		Email:        req.Email,
		Provider:     req.Provider,
		Protocol:     req.Protocol,
		Server:       req.Server,
		Port:         req.Port,
		UseTLS:       req.UseTLS,
		AuthType:     req.AuthType,
		AccessToken:  req.AccessToken,
		RefreshToken: req.RefreshToken,
		TokenExpiry:  req.TokenExpiry,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, httpdto.ErrorOutput{Message: err.Error()})
	}
	return c.JSON(http.StatusCreated, httpdto.MailboxResponse{ID: mb.ID})
}

func (h *MailboxHandler) TriggerSync(c echo.Context) error {
	id := c.Param("id")
	ctx := c.Request().Context()

	err := h.syncService.SyncMailbox(ctx, id)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return c.JSON(http.StatusNotFound, httpdto.ErrorOutput{Message: "mailbox not found"})
		}
		return c.JSON(http.StatusInternalServerError, httpdto.ErrorOutput{Message: err.Error()})
	}
	return c.NoContent(http.StatusAccepted)
}
