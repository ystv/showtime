package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handlers) listYouTubeBroadcasts(c echo.Context) error {
	broadcasts, err := h.yt.GetBroadcasts(c.Request().Context())
	if err != nil {
		err = fmt.Errorf("failed to get streams: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, broadcasts)
}

func (h *Handlers) enableYouTube(c echo.Context) error {
	h.yt.EnableShowTimeForBroadcast(c.Request().Context(), c.Param("broadcastID"))
	return c.NoContent(http.StatusOK)
}

func (h *Handlers) disableYouTube(c echo.Context) error {
	h.yt.DisableShowTimeForBroadcast(c.Request().Context(), c.Param("broadcastID"))
	return c.NoContent(http.StatusOK)
}
