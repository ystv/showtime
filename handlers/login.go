package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"
)

func (h *Handlers) googleLogin(w http.ResponseWriter, r *http.Request) {
	state := h.generateStateOauthCookie(w)
	url := h.auth.GetAuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
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
