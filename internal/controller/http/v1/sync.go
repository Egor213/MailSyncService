package httpapi

import (
	"mail-sync-service/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

type SyncHandler struct {
	syncService service.Sync
}

func newSyncRoutes(g *echo.Group, syncService service.Sync) {
	h := &SyncHandler{syncService: syncService}

	g.POST("/mailbox/:id", h.SyncNow)      // POST /api/v1/sync/mailbox/:id
	g.GET("/mailbox/:id/status", h.Status) // GET /api/v1/sync/mailbox/:id/status
}

func (h *SyncHandler) SyncNow(c echo.Context) error {
	id := c.Param("id")
	if err := h.syncService.SyncMailbox(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.NoContent(http.StatusAccepted)
}

func (h *SyncHandler) Status(c echo.Context) error {
	id := c.Param("id")
	job, err := h.syncService.GetLastSyncStatus(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, job)
}
