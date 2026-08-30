package httpapi

import (
	"mail-sync-service/internal/service"

	"github.com/labstack/echo/v4"
)

type SyncHandler struct {
	syncService *service.SyncService
}

func newSyncRoutes(g *echo.Group, syncService *service.SyncService) {
	h := &SyncHandler{syncService: syncService}

	g.POST("/mailbox/:id", h.SyncNow)      // POST /api/v1/sync/mailbox/:id
	g.GET("/mailbox/:id/status", h.Status) // GET /api/v1/sync/mailbox/:id/status
}

func (h *SyncHandler) SyncNow(c echo.Context) error {
	// ... запуск синхронизации через Kafka или напрямую
	return nil
}

func (h *SyncHandler) Status(c echo.Context) error {
	// ... возвращает последний статус синхронизации
	return nil
}
