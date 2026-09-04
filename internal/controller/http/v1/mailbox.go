package httpapi

import (
	"errors"
	httpdto "mail-sync-service/internal/controller/http/v1/dto"
	"mail-sync-service/internal/entity"
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

	g.POST("", h.Create)               // POST   /api/v1/mailboxes
	g.GET("", h.ListActive)            // GET    /api/v1/mailboxes
	g.GET("/:id", h.Get)              // GET    /api/v1/mailboxes/:id
	g.PUT("/:id", h.Update)           // PUT    /api/v1/mailboxes/:id
	g.DELETE("/:id", h.Delete)        // DELETE /api/v1/mailboxes/:id
	g.POST("/:id/sync", h.TriggerSync) // POST   /api/v1/mailboxes/:id/sync
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

func (h *MailboxHandler) ListActive(c echo.Context) error {
	mailboxes, err := h.mailboxService.ListActive(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, httpdto.ErrorOutput{Message: err.Error()})
	}

	resp := make([]httpdto.MailboxResponse, 0, len(mailboxes))
	for _, mb := range mailboxes {
		resp = append(resp, mailboxToResponse(mb))
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *MailboxHandler) Get(c echo.Context) error {
	id := c.Param("id")
	mb, err := h.mailboxService.GetMailbox(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return c.JSON(http.StatusNotFound, httpdto.ErrorOutput{Message: "mailbox not found"})
		}
		return c.JSON(http.StatusInternalServerError, httpdto.ErrorOutput{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, mailboxToResponse(mb))
}

func (h *MailboxHandler) Update(c echo.Context) error {
	id := c.Param("id")

	var req httpdto.UpdateMailboxRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httpdto.ErrorOutput{Message: "invalid request"})
	}

	mb, err := h.mailboxService.GetMailbox(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return c.JSON(http.StatusNotFound, httpdto.ErrorOutput{Message: "mailbox not found"})
		}
		return c.JSON(http.StatusInternalServerError, httpdto.ErrorOutput{Message: err.Error()})
	}

	if req.Email != "" {
		mb.Email = req.Email
	}
	if req.Server != "" {
		mb.Server = req.Server
	}
	if req.Port != 0 {
		mb.Port = req.Port
	}
	if req.UseTLS != nil {
		mb.UseTLS = *req.UseTLS
	}
	if req.AccessToken != "" {
		mb.AccessToken = req.AccessToken
	}
	if req.RefreshToken != "" {
		mb.RefreshToken = req.RefreshToken
	}
	if req.TokenExpiry != nil {
		mb.TokenExpiry = req.TokenExpiry
	}
	if req.IsActive != nil {
		mb.IsActive = *req.IsActive
	}

	if err := h.mailboxService.UpdateMailbox(c.Request().Context(), mb); err != nil {
		return c.JSON(http.StatusInternalServerError, httpdto.ErrorOutput{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, mailboxToResponse(mb))
}

func (h *MailboxHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.mailboxService.DeleteMailbox(c.Request().Context(), id); err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return c.JSON(http.StatusNotFound, httpdto.ErrorOutput{Message: "mailbox not found"})
		}
		return c.JSON(http.StatusInternalServerError, httpdto.ErrorOutput{Message: err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
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

func mailboxToResponse(mb *entity.Mailbox) httpdto.MailboxResponse {
	return httpdto.MailboxResponse{
		ID:          mb.ID,
		Email:       mb.Email,
		ProviderID:  mb.ProviderID,
		ProtocolID:  mb.ProtocolID,
		Server:      mb.Server,
		Port:        mb.Port,
		UseTLS:      mb.UseTLS,
		AuthTypeID:  mb.AuthTypeID,
		TokenExpiry: mb.TokenExpiry,
		CreatedAt:   mb.CreatedAt,
		UpdatedAt:   mb.UpdatedAt,
		IsActive:    mb.IsActive,
	}
}
