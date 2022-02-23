package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
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
