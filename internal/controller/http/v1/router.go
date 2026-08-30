package httpapi

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"mail-sync-service/internal/service"
	errutils "mail-sync-service/pkg/errors"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func ConfigureRouter(handler *echo.Echo, services *service.Services) {
	// Логирование
	logFile := setLogsFile()
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	handler.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Output: multiWriter,
	}))
	handler.Use(middleware.Recover())

	// Healthcheck
	handler.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	api := handler.Group("/api/v1")
	{
		newOAuthRoutes(api.Group("/oauth"), services.OAuth)
	}
	// Можно добавить middleware для аутентификации, пока пусто
	// authMW := mw.NewAuth(services.Auth)

	// newMailboxRoutes(api.Group("/mailboxes"), services.Mailbox)
	// newSyncRoutes(api.Group("/sync"), services.Sync)
}

func setLogsFile() *os.File {
	logPath := filepath.Join("logs", "logfile.log")
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		log.Fatal(errutils.WrapPathErr(err))
	}
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0o666)
	if err != nil {
		log.Fatal(errutils.WrapPathErr(err))
	}
	return file
}
