package handlers

import (
	"fmt"
	"net/http"

	"github.com/ystv/showtime/youtube"
)

func (h *Handlers) showStreams(w http.ResponseWriter, r *http.Request) {
	tok, err := h.auth.GetToken("me")
	if err != nil {
		err = fmt.Errorf("failed to get token: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	yt, err := youtube.New(h.auth.GetClient(r.Context(), tok))
	if err != nil {
		err = fmt.Errorf("failed to create youtube service: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = yt.GetStreams(r.Context())
	if err != nil {
		err = fmt.Errorf("failed to get streams: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
