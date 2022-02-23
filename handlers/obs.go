package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handlers) obsListPlayouts(c echo.Context) error {
	po, err := h.play.List(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.Render(http.StatusOK, "list-playouts", po)
}
