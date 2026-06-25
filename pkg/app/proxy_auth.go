package app

import (
	"errors"
	"net/mail"
	"net/http"
	"strings"

	"github.com/cat-dealer/go-rand/v2"
	"github.com/jovandeginste/workout-tracker/v2/pkg/database"
	"github.com/labstack/echo/v4"
)

const (
	lazycatUserIDHeader   = "X-HC-User-ID"
	lazycatUserRoleHeader = "X-HC-User-Role"
	lazycatAdminRole      = "ADMIN"
)

func (a *App) ProxyAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !a.Config.ProxyAuthEnabled {
			return next(c)
		}

		uid := strings.TrimSpace(c.Request().Header.Get(lazycatUserIDHeader))
		if uid == "" || a.hasRequestAuth(c) {
			return next(c)
		}

		if err := a.loginProxyAuthUser(c, uid, c.Request().Header.Get(lazycatUserRoleHeader)); err != nil {
			a.logger.Warn("proxy auth login failed", "error", err, "uid", uid)
		}

		return next(c)
	}
}

func (a *App) loginProxyAuthUser(c echo.Context, uid, role string) error {
	u, err := a.findOrCreateProxyAuthUser(uid, role)
	if err != nil {
		return err
	}

	a.sessionManager.Put(c.Request().Context(), "username", u.Username)

	return a.createToken(u, c)
}

func (a *App) findOrCreateProxyAuthUser(uid, role string) (*database.User, error) {
	username, err := normalizeProxyAuthUsername(uid)
	if err != nil {
		return nil, err
	}

	if u, err := database.GetUser(a.db, username); err == nil {
		return u, nil
	}

	u := &database.User{
		UserData: database.UserData{
			Username: username,
			Name:     uid,
			Active:   true,
			Admin:    strings.EqualFold(strings.TrimSpace(role), lazycatAdminRole),
		},
	}
	u.Profile.Theme = BrowserTheme
	u.Profile.TotalsShow = DefaultTotalsShow
	u.Profile.Language = DefaultLanguage
	u.Profile.User = u

	password := rand.String(32, rand.GetAlphaNumericPool())
	if err := u.SetPassword(password); err != nil {
		return nil, err
	}

	if err := u.Create(a.db); err != nil {
		if existing, getErr := database.GetUser(a.db, username); getErr == nil {
			return existing, nil
		}

		return nil, err
	}

	return database.GetUser(a.db, username)
}

func normalizeProxyAuthUsername(uid string) (string, error) {
	uid = strings.TrimSpace(uid)
	if uid == "" {
		return "", errors.New("empty proxy auth uid")
	}

	if len(uid) > database.UsernameMaximumLength {
		uid = uid[:database.UsernameMaximumLength]
	}

	if isValidProxyAuthUsername(uid) {
		return uid, nil
	}

	fallback := sanitizeProxyAuthUsername(uid)
	if !isValidProxyAuthUsername(fallback) {
		return "", database.ErrUsernameInvalid
	}

	return fallback, nil
}

func sanitizeProxyAuthUsername(uid string) string {
	var local strings.Builder

	for _, r := range uid {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '_', r == '-':
			local.WriteRune(r)
		default:
			local.WriteRune('_')
		}
	}

	clean := strings.Trim(local.String(), "._-")
	if clean == "" {
		clean = "user"
	}

	if len(clean) > 20 {
		clean = clean[:20]
	}

	return clean + "@lzc.local"
}

func isValidProxyAuthUsername(username string) bool {
	if len(username) < database.UsernameMinimumLength || len(username) > database.UsernameMaximumLength {
		return false
	}

	if _, err := mail.ParseAddress(username); err == nil {
		return true
	}

	_, err := mail.ParseAddress(username + "@localhost")

	return err == nil
}

func (a *App) hasRequestAuth(c echo.Context) bool {
	if username, ok := a.sessionManager.Get(c.Request().Context(), "username").(string); ok && username != "" {
		return true
	}

	cookie, err := c.Cookie("token")

	return err == nil && cookie != nil && cookie.Value != ""
}

func (a *App) redirectIfProxyAuthenticated(c echo.Context) error {
	if !a.Config.ProxyAuthEnabled {
		return nil
	}

	uid := strings.TrimSpace(c.Request().Header.Get(lazycatUserIDHeader))
	if uid == "" {
		return nil
	}

	if !a.hasRequestAuth(c) {
		if err := a.loginProxyAuthUser(c, uid, c.Request().Header.Get(lazycatUserRoleHeader)); err != nil {
			a.logger.Warn("proxy auth redirect failed", "error", err, "uid", uid)

			return nil
		}
	}

	return c.Redirect(http.StatusFound, a.redirectAfterLoginTarget(c))
}
