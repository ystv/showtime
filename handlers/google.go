package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func (h *Handlers) loginGoogle(c echo.Context) error {
	state := h.generateStateOauthCookie(c.Response().Writer)
	url := h.auth.GetAuthCodeURL(state)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *Handlers) generateStateOauthCookie(w http.ResponseWriter) string {
	expiration := time.Now().Add(15 * time.Minute)

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: h.conf.StateCookieName, Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}

func (h *Handlers) callbackGoogle(c echo.Context) error {
	// Check state cookie to make sure there isn't any CSRF biz
	state, _ := c.Cookie(h.conf.StateCookieName)

	if c.FormValue("state") != state.Value {
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	code := c.FormValue("code")
	if code == "" {
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	tok, err := h.auth.NewToken(c.Request().Context(), code)
	if err != nil {
		err = fmt.Errorf("failed to get token: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	err = h.auth.StoreToken("me", tok)
	if err != nil {
		err = fmt.Errorf("failed to store token: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.String(http.StatusOK, "login successful!")
}
