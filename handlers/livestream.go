package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/ystv/showtime/livestream"
)

func (h *Handlers) newLivestream(c echo.Context) error {
	strm := livestream.EditLivestream{}
	err := c.Bind(&strm)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	strmID, err := h.ls.New(c.Request().Context(), strm)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusCreated, strmID)
}

func (h *Handlers) listLivestreams(c echo.Context) error {
	strms, err := h.ls.List(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, strms)
}

func (h *Handlers) getLivestreamEvents(c echo.Context) error {
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	evts, err := h.ls.ListEvents(c.Request().Context(), strmID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	if evts == nil {
		return c.JSON(http.StatusOK, []string{})
	}
	return c.JSON(http.StatusOK, evts)
}

func (h *Handlers) updateLivestream(c echo.Context) error {
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	strm := livestream.EditLivestream{}
	err = c.Bind(&strm)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	err = h.ls.Update(c.Request().Context(), strmID, strm)
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
