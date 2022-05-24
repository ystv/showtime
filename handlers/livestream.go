package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/ystv/showtime/livestream"
)

func (h *Handlers) newLivestream(c echo.Context) error {
	strm := livestream.NewLivestream{}
	err := c.Bind(&strm)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	err = h.ls.New(c.Request().Context(), strm)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return nil
}

func (h *Handlers) listLivestreams(c echo.Context) error {
	strms, err := h.ls.List(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, strms)
}

func (h *Handlers) updateLivestream(c echo.Context) error {
	strm := livestream.Livestream{}
	err := c.Bind(&strm)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	err = h.ls.Update(c.Request().Context(), strm)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return nil
}

func (h *Handlers) refreshStreamKey(c echo.Context) error {
	err := h.ls.RefreshStreamKey(c.Request().Context(), c.Param("playoutID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.NoContent(http.StatusOK)
}
