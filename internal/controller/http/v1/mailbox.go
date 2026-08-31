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
	syncService    service.SyncService
}

func newMailboxRoutes(g *echo.Group, mailboxService *service.MailboxService) {
	h := &MailboxHandler{mailboxService: mailboxService}

	// g.POST("", h.Create)               // POST /api/v1/mailboxes
	// g.GET("/:id", h.Get)               // GET /api/v1/mailboxes/:id
	// g.PUT("/:id", h.Update)            // PUT /api/v1/mailboxes/:id
	// g.DELETE("/:id", h.Delete)         // DELETE /api/v1/mailboxes/:id
	g.POST("/:id/sync", h.TriggerSync) // POST /api/v1/mailboxes/:id/sync
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

// func (h *MailboxHandler) Create(c echo.Context) error {
// 	var req dto.CreateMailboxRequest
// 	if err := c.Bind(&req); err != nil {
// 		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "invalid request"})
// 	}
// 	if err := c.Validate(req); err != nil {
// 		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: err.Error()})
// 	}
// 	mb := req.ToEntity()
// 	if err := h.mailboxService.CreateMailbox(c.Request().Context(), mb); err != nil {
// 		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
// 	}
// 	return c.JSON(http.StatusCreated, dto.MailboxResponse{ID: mb.ID})
// }

// func (h *MailboxHandler) Sync(c echo.Context) error {
// 	id := c.Param("id")
// 	if err := h.mailboxService.TriggerSync(c.Request().Context(), id); err != nil {
// 		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
// 	}
// 	return c.NoContent(http.StatusAccepted)
// }
