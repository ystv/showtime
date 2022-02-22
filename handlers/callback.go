package handlers

import (
	"fmt"
	"log"
	"net/http"
)

func (h *Handlers) googleCallback(w http.ResponseWriter, r *http.Request) {
	// Check state cookie to make sure there isn't any CSRF biz
	state, _ := r.Cookie(h.stateCookieName)

	if r.FormValue("state") != state.Value {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	if code == "" {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	tok, err := h.auth.NewToken(r.Context(), code)
	if err != nil {
		log.Printf("failed to get token: %+v", err)
	}

	h.auth.StoreToken("me", tok)
	fmt.Fprintf(w, "successful login!")
}
