package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jovandeginste/workout-tracker/v2/pkg/database"
	session "github.com/spazzymoto/echo-scs-session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeProxyAuthUsername(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		uid     string
		want    string
		wantErr bool
	}{
		{name: "plain uid", uid: "alice", want: "alice"},
		{name: "email uid", uid: "alice@example.com", want: "alice@example.com"},
		{name: "invalid chars", uid: "bad uid!", want: "bad_uid@lzc.local"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := normalizeProxyAuthUsername(tt.uid)
			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProxyAuthMiddleware_LogsInFromHeader(t *testing.T) {
	t.Setenv("WT_DATABASE_DRIVER", "memory")
	t.Setenv("WT_PROXY_AUTH_ENABLED", "true")

	a := defaultApp(t)
	require.NoError(t, a.Configure())

	e := a.echo
	req := httptest.NewRequest(http.MethodGet, e.Reverse("user-login"), nil)
	req.Header.Set(lazycatUserIDHeader, "lazycat-user")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), e.Reverse("dashboard"))

	var tokenCookie *http.Cookie
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == "token" {
			tokenCookie = cookie
		}
	}

	require.NotNil(t, tokenCookie)
	assert.NotEmpty(t, tokenCookie.Value)

	u, err := database.GetUser(a.db, "lazycat-user")
	require.NoError(t, err)
	assert.True(t, u.Active)
}

func TestProxyAuthMiddleware_DisabledWithoutConfig(t *testing.T) {
	t.Setenv("WT_DATABASE_DRIVER", "memory")

	a := defaultApp(t)
	require.NoError(t, a.Configure())

	e := a.echo
	req := httptest.NewRequest(http.MethodGet, e.Reverse("user-login"), nil)
	req.Header.Set(lazycatUserIDHeader, "lazycat-user")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Sign in")
}

func TestUserLoginHandler_ProxyAuthRedirect(t *testing.T) {
	t.Setenv("WT_DATABASE_DRIVER", "memory")
	t.Setenv("WT_PROXY_AUTH_ENABLED", "true")

	a := configuredApp(t)
	e := a.echo

	req := httptest.NewRequest(http.MethodGet, e.Reverse("user-login"), nil)
	req.Header.Set(lazycatUserIDHeader, "another-user")
	rec := httptest.NewRecorder()

	s := session.LoadAndSave(a.sessionManager)
	e.ServeHTTP(rec, req)

	_ = s

	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), e.Reverse("dashboard"))
}
