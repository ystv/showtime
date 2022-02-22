package handlers

import (
	"fmt"
	"net/http"

	"github.com/ystv/showtime/auth"
)

type Handlers struct {
	auth            *auth.Auther
	mux             *http.ServeMux
	stateCookieName string
}

func New(auth *auth.Auther) *Handlers {
	return &Handlers{
		auth:            auth,
		mux:             http.NewServeMux(),
		stateCookieName: "state-token",
	}
}

func (h *Handlers) GetHandlers() http.Handler {
	h.mux.HandleFunc("/", h.index)
	h.mux.HandleFunc("/streams", h.showStreams)
	h.mux.HandleFunc("/oauth/google/login", h.googleLogin)
	h.mux.HandleFunc("/oauth/google/callback", h.googleCallback)

	return h.mux
}

func (h *Handlers) index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "it's show time!")
}
