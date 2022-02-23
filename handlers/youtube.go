package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handlers) listYouTubeStreams(c echo.Context) error {
	broadcasts, err := h.yt.GetBroadcasts(c.Request().Context())
	if err != nil {
		err = fmt.Errorf("failed to get streams: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, broadcasts)
}
