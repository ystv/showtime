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

func (h *Handlers) obsNewPlayout(c echo.Context) error {
	return c.Render(http.StatusOK, "new-playout", nil)
}

func (h *Handlers) obsNewPlayoutSubmit(c echo.Context) error {
	po := playout.NewPlayout{}
	err := c.Bind(&po)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	err = h.play.New(c.Request().Context(), po)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return h.obsListPlayouts(c)
}

func (h *Handlers) obsEndPlayout(c echo.Context) error {
	err := h.play.End(c.Request().Context(), c.Param("playoutID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return h.obsListPlayouts(c)
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

func (h *Handlers) obsLinkToPublicSite(c echo.Context) error {
	playoutID, err := strconv.Atoi(c.Param("playoutID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	p, err := h.play.Get(c.Request().Context(), playoutID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	data := struct {
		Playout  playout.Playout
		Channels []channel.Channel
	}{
		Playout: p,
	}
	return c.Render(http.StatusOK, "set-public-site-link", data)
}

func (h *Handlers) obsLinkToPublicSiteConfirm(c echo.Context) error {
	return c.Render(http.StatusCreated, "successful-link", c.Param("playoutID"))
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

func (h *Handlers) obsLinkToYouTubeConfirm(c echo.Context) error {
	err := h.yt.EnableShowTimeForBroadcast(c.Request().Context(), c.FormValue("broadcastID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	err = h.play.UpdateYouTubeLink(c.Request().Context(), c.Param("playoutID"), c.FormValue("broadcastID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.Render(http.StatusCreated, "successful-link", c.Param("playoutID"))
}

func (h *Handlers) obsUnlinkFromYouTube(c echo.Context) error {
	err := h.yt.DisableShowTimeForBroadcast(c.Request().Context(), c.Param("broadcastID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	err = h.play.UpdateYouTubeLink(c.Request().Context(), c.Param("playoutID"), "")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.Render(http.StatusOK, "successful-unlink", c.Param("playoutID"))
}
