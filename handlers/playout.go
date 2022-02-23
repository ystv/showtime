package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/ystv/showtime/playout"
)

func (h *Handlers) newPlayout(c echo.Context) error {
	po := playout.NewPlayout{}
	err := c.Bind(&po)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	err = h.play.New(c.Request().Context(), po)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return nil
}

func (h *Handlers) listPlayouts(c echo.Context) error {
	po, err := h.play.List(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, po)
}

func (h *Handlers) updatePlayout(c echo.Context) error {
	po := playout.Playout{}
	err := c.Bind(&po)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	err = h.play.Update(c.Request().Context(), po)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return nil
}
