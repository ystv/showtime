package handlers

import (
	"log"
	"net/http"
)

func (h *Handlers) showVideos(w http.ResponseWriter, r *http.Request) {
	_, err := h.auth.GetToken("me")
	if err != nil {
		log.Printf("failed to get token: %+v", err)
	}

}
