package httpapi

import (
	"mail-sync-service/internal/service"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type MetricsHandler struct {
	metricsService service.Metrics
}

func newMetricsRoutes(g *echo.Group, metricsService service.Metrics) {
	h := &MetricsHandler{metricsService: metricsService}

	g.GET("/overview", h.Overview)
}

func (h *MetricsHandler) Overview(c echo.Context) error {
	since := time.Now().Add(-24 * time.Hour)
	if sinceStr := c.QueryParam("since"); sinceStr != "" {
		if t, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			since = t
		}
	}
	stats, err := h.metricsService.GetOverviewStats(c.Request().Context(), since)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, stats)
}
