package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func (h *Handlers) googleLogin(c echo.Context) error {
	state := h.generateStateOauthCookie(c.Response().Writer)
	url := h.auth.GetAuthCodeURL(state)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *Handlers) generateStateOauthCookie(w http.ResponseWriter) string {
	expiration := time.Now().Add(15 * time.Minute)

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: h.stateCookieName, Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}
