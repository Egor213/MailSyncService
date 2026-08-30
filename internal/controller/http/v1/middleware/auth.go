package middleware

import (
	"github.com/labstack/echo/v4"
)

// AuthMiddleware — пример (можно расширить)
type AuthMiddleware struct{}

func NewAuth() *AuthMiddleware {
	return &AuthMiddleware{}
}

func (m *AuthMiddleware) UserIdentity(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// проверка токена, установка контекста
		return next(c)
	}
}

func (m *AuthMiddleware) CheckRole(role string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// проверка роли
			return next(c)
		}
	}
}
