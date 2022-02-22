package handlers

import (
	"encoding/json"
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
	broadcasts, err := yt.GetBroadcasts(r.Context())
	if err != nil {
		err = fmt.Errorf("failed to get streams: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(broadcasts)
	if err != nil {
		err = fmt.Errorf("failed to encode to json: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
