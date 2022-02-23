package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/ystv/showtime/playout"
	"github.com/ystv/showtime/youtube"
)

func (h *Handlers) obsListPlayouts(c echo.Context) error {
	po, err := h.play.List(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.Render(http.StatusOK, "list-playouts", po)
}

func (h *Handlers) obsGetPlayout(c echo.Context) error {
	po, err := h.play.List(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	for _, playout := range po {
		if strconv.Itoa(playout.PlayoutID) == c.Param("playoutID") {
			return c.Render(http.StatusOK, "get-playout", playout)
		}
	}
	return echo.NewHTTPError(http.StatusNotFound)
}

func (h *Handlers) obsManagePlayout(c echo.Context) error {
	po, err := h.play.List(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	for _, playout := range po {
		if strconv.Itoa(playout.PlayoutID) == c.Param("playoutID") {
			return c.Render(http.StatusOK, "manage-playout", playout)
		}
	}
	return echo.NewHTTPError(http.StatusNotFound)
}

func (h *Handlers) obsLinkToYouTube(c echo.Context) error {
	po, err := h.play.List(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	for _, p := range po {
		if strconv.Itoa(p.PlayoutID) == c.Param("playoutID") {
			broadcasts, err := h.yt.GetBroadcasts(c.Request().Context())
			if err != nil {
				return fmt.Errorf("failed to get youtube broadcasts: %w", err)
			}
			data := struct {
				Playout    playout.Playout
				Broadcasts []youtube.Broadcast
			}{
				Playout:    p,
				Broadcasts: broadcasts,
			}
			return c.Render(http.StatusOK, "set-youtube-link", data)
		}
	}
	return echo.NewHTTPError(http.StatusNotFound)
}
