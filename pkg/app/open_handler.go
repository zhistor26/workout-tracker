package app

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jovandeginste/workout-tracker/v2/pkg/converters"
	"github.com/jovandeginste/workout-tracker/v2/pkg/database"
	"github.com/labstack/echo/v4"
)

var (
	ErrOpenMissingFile = errors.New("missing file parameter")
	ErrOpenFileNotFound = errors.New("file not found")
)

func resolveLazyCatFilePath(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", fmt.Errorf("%w", ErrOpenMissingFile)
	}

	path := strings.TrimSpace(raw)
	if decoded, err := url.QueryUnescape(raw); err == nil && strings.TrimSpace(decoded) != "" {
		path = strings.TrimSpace(decoded)
	}

	path = strings.TrimPrefix(path, "file://")

	candidates := []string{path}

	if strings.HasPrefix(path, "/home/") {
		rest := strings.TrimPrefix(path, "/home/")
		if uid, fileRest, ok := strings.Cut(rest, "/"); ok && fileRest != "" {
			candidates = append(candidates,
				filepath.Join("/lzcapp/documents", uid, fileRest),
				filepath.Join("/lzcapp/run/mnt/home", uid, fileRest),
				filepath.Join("/lzcapp/document", uid, fileRest),
			)
		}
	}

	if idx := strings.Index(path, "/_lzc/files/home/"); idx >= 0 {
		rest := strings.TrimPrefix(path[idx:], "/_lzc/files/home/")
		if uid, fileRest, ok := strings.Cut(rest, "/"); ok && fileRest != "" {
			candidates = append(candidates,
				filepath.Join("/lzcapp/documents", uid, fileRest),
				filepath.Join("/lzcapp/run/mnt/home", uid, fileRest),
			)
		}
	}

	if !strings.HasPrefix(path, "/") {
		candidates = append(candidates, filepath.Join("/lzcapp/run/mnt/home", path))
	}

	seen := map[string]struct{}{}

	for _, candidate := range candidates {
		candidate = filepath.Clean(candidate)
		if _, ok := seen[candidate]; ok {
			continue
		}

		seen[candidate] = struct{}{}

		info, err := os.Stat(candidate)
		if err != nil || info.IsDir() {
			continue
		}

		return candidate, nil
	}

	return "", fmt.Errorf("%w: %q", ErrOpenFileNotFound, raw)
}

func (a *App) openHandler(c echo.Context) error {
	if err := a.ensureOpenSession(c); err != nil {
		return err
	}

	u, ok := a.loadUserFromRequestToken(c)
	if !ok || u.IsAnonymous() {
		return a.redirectOpenToLogin(c)
	}

	rawFile := c.QueryParam("file")
	if rawFile == "" {
		return a.redirectWithError(c, a.echo.Reverse("workouts"), ErrOpenMissingFile)
	}

	filePath, err := resolveLazyCatFilePath(rawFile)
	if err != nil {
		a.logger.Warn("open handler: resolve file failed", "file", rawFile, "error", err)

		return a.redirectWithError(c, a.echo.Reverse("workouts"), err)
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	if !slices.Contains(converters.SupportedFileTypes, ext) {
		return a.redirectWithError(c, a.echo.Reverse("workouts"), fmt.Errorf("unsupported file type: %s", ext))
	}

	dat, err := os.ReadFile(filePath)
	if err != nil {
		a.logger.Warn("open handler: read file failed", "path", filePath, "error", err)

		return a.redirectWithError(c, a.echo.Reverse("workouts"), err)
	}

	workouts, addErrs := u.AddWorkout(a.db, database.WorkoutTypeAutoDetect, "", filePath, dat)
	if len(addErrs) > 0 {
		return a.redirectWithError(c, a.echo.Reverse("workouts"), addErrs[0])
	}

	if len(workouts) == 0 {
		return a.redirectWithError(c, a.echo.Reverse("workouts"), ErrNothingImported)
	}

	a.addNoticeT(c, "translation.The_workout_s_has_been_created", workouts[0].Name)

	return c.Redirect(http.StatusFound, a.echo.Reverse("workout-show", workouts[0].ID))
}

func (a *App) ensureOpenSession(c echo.Context) error {
	if _, ok := a.loadUserFromRequestToken(c); ok {
		return nil
	}

	uid := strings.TrimSpace(c.Request().Header.Get(lazycatUserIDHeader))
	if a.Config.ProxyAuthEnabled && uid != "" {
		if err := a.loginProxyAuthUser(c, uid, c.Request().Header.Get(lazycatUserRoleHeader)); err != nil {
			a.logger.Warn("open handler proxy auth failed", "error", err, "uid", uid)
		} else {
			return c.Redirect(http.StatusFound, c.Request().URL.String())
		}
	}

	return a.redirectOpenToLogin(c)
}

func (a *App) redirectOpenToLogin(c echo.Context) error {
	a.rememberRedirectAfterLogin(c, c.Request().URL.RequestURI())

	return c.Redirect(http.StatusFound, a.echo.Reverse("user-login"))
}

func (a *App) loadUserFromRequestToken(c echo.Context) (*database.User, bool) {
	cookie, err := c.Cookie("token")
	if err != nil || cookie == nil || cookie.Value == "" {
		return nil, false
	}

	token, err := jwt.ParseWithClaims(cookie.Value, jwt.MapClaims{}, func(t *jwt.Token) (any, error) {
		return a.jwtSecret(), nil
	})
	if err != nil {
		return nil, false
	}

	c.Set("user", token)

	if err := a.setUserFromContext(c); err != nil {
		return nil, false
	}

	return a.getCurrentUser(c), true
}
