package app

import (
	"strings"

	"github.com/labstack/echo/v4"
)

const redirectAfterLoginSessionKey = "redirect_after_login"

func isSafeRedirectPath(p string) bool {
	if p == "" || !strings.HasPrefix(p, "/") || strings.HasPrefix(p, "//") {
		return false
	}

	return true
}

func (a *App) redirectAfterLoginTarget(c echo.Context) string {
	if next := c.QueryParam("next"); isSafeRedirectPath(next) {
		return next
	}

	if next, ok := a.sessionManager.Get(c.Request().Context(), redirectAfterLoginSessionKey).(string); ok && isSafeRedirectPath(next) {
		a.sessionManager.Remove(c.Request().Context(), redirectAfterLoginSessionKey)

		return next
	}

	return a.echo.Reverse("dashboard")
}

func (a *App) rememberRedirectAfterLogin(c echo.Context, target string) {
	if !isSafeRedirectPath(target) {
		return
	}

	a.sessionManager.Put(c.Request().Context(), redirectAfterLoginSessionKey, target)
}
