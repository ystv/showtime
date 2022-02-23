package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/ystv/showtime/playout"
)

func (h *Handlers) hookStreamStart(c echo.Context) error {
	c.Request().ParseForm()
	streamKey := c.Request().FormValue("name")

	log.Println(streamKey)

	_, err := h.play.GetByStreamKey(c.Request().Context(), streamKey)
	if err != nil {
		if errors.Is(err, playout.ErrStreamKeyNotFound) {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.NoContent(http.StatusOK)
}
