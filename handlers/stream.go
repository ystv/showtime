package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/ystv/showtime/youtube"
)

func (h *Handlers) showStreams(c echo.Context) error {
	tok, err := h.auth.GetToken("me")
	if err != nil {
		err = fmt.Errorf("failed to get token: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	yt, err := youtube.New(h.auth.GetClient(c.Request().Context(), tok))
	if err != nil {
		err = fmt.Errorf("failed to create youtube service: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	broadcasts, err := yt.GetBroadcasts(c.Request().Context())
	if err != nil {
		err = fmt.Errorf("failed to get streams: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, broadcasts)
}
