package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/ystv/showtime/livestream"
)

func (h *Handlers) hookStreamStart(c echo.Context) error {
	if err := c.Request().ParseForm(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	streamKey := c.Request().FormValue("name")

	strm, err := h.ls.GetByStreamKey(c.Request().Context(), streamKey)
	if err != nil {
		if errors.Is(err, livestream.ErrStreamKeyNotFound) {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	err = h.ls.Forward(c.Request().Context(), strm)
	if err != nil {
		log.Printf("failed to forward stream %d: %v", strm.ID, err)
		_ = h.ls.CreateEvent(c.Request().Context(), strm.ID, livestream.EventError, livestream.EventErrorPayload{
			Err:     err.Error(),
			Context: "ls.Forward",
		})
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	err = h.ls.CreateEvent(c.Request().Context(), strm.ID, livestream.EventStreamReceived, livestream.EventStreamReceivedPayload{})
	if err != nil {
		log.Printf("failed to create stream start event: %v", err)
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handlers) hookStreamDone(c echo.Context) error {
	if err := c.Request().ParseForm(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	streamKey := c.Request().FormValue("name")

	strm, err := h.ls.GetByStreamKey(c.Request().Context(), streamKey)
	if err != nil {
		if errors.Is(err, livestream.ErrStreamKeyNotFound) {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	err = h.ls.CreateEvent(c.Request().Context(), strm.ID, livestream.EventStreamLost, livestream.EventStreamLostPayload{})
	if err != nil {
		log.Printf("failed to create stream done event: %v", err)
	}

	return c.NoContent(http.StatusOK)
}
