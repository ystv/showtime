package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/ystv/showtime/livestream"
	"github.com/ystv/showtime/mcr"
	"github.com/ystv/showtime/youtube"
)

func (h *Handlers) obsHome(c echo.Context) error {
	return c.Render(http.StatusOK, "home", nil)
}

func (h *Handlers) obsListLivestreams(c echo.Context) error {
	strms, err := h.ls.List(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.Render(http.StatusOK, "list-livestreams", strms)
}

func (h *Handlers) obsGetLivestream(c echo.Context) error {
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	strm, err := h.ls.Get(c.Request().Context(), strmID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.Render(http.StatusOK, "get-livestream", strm)
}

func (h *Handlers) obsNewLivestream(c echo.Context) error {
	return c.Render(http.StatusOK, "new-livestream", nil)
}

func (h *Handlers) obsNewLivestreamSubmit(c echo.Context) error {
	strm := livestream.NewLivestream{}
	err := c.Bind(&strm)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	err = h.ls.New(c.Request().Context(), strm)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return h.obsListLivestreams(c)
}

func (h *Handlers) obsStartLivestream(c echo.Context) error {
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	err = h.ls.Start(c.Request().Context(), strmID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return h.obsGetLivestream(c)
}

func (h *Handlers) obsEndLivestream(c echo.Context) error {
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	err = h.ls.End(c.Request().Context(), strmID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return h.obsListLivestreams(c)
}

func (h *Handlers) obsManageLivestream(c echo.Context) error {
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	strm, err := h.ls.Get(c.Request().Context(), strmID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.Render(http.StatusOK, "manage-livestream", strm)
}

func (h *Handlers) obsLinkToMCR(c echo.Context) error {
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	strm, err := h.ls.Get(c.Request().Context(), strmID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	ch, err := h.mcr.ListChannels(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	data := struct {
		Livestream livestream.Livestream
		Channels   []mcr.Channel
	}{
		Livestream: strm,
		Channels:   ch,
	}
	return c.Render(http.StatusOK, "set-mcr-link", data)
}

func (h *Handlers) obsLinkToMCRConfirm(c echo.Context) error {
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	strm, err := h.ls.Get(c.Request().Context(), strmID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	res := struct {
		ChannelID string `form:"channelID"`
	}{}
	err = c.Bind(&res)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	po := mcr.NewPlayout{
		ChannelID:  res.ChannelID,
		SrcURI:     h.conf.IngestAddress + "/" + strm.StreamKey,
		Title:      strm.Title,
		Visibility: "public",
	}
	playoutID, err := h.mcr.NewPlayout(c.Request().Context(), po)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	err = h.ls.UpdateMCRLink(c.Request().Context(), strmID, strconv.Itoa(playoutID))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.Render(http.StatusCreated, "successful-link", c.Param("livestreamID"))
}

func (h *Handlers) obsUnlinkFromMCR(c echo.Context) error {
	linkID, err := strconv.Atoi(c.Param("linkID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	err = h.mcr.DeletePlayout(c.Request().Context(), linkID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	err = h.ls.UpdateMCRLink(c.Request().Context(), strmID, "")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.Render(http.StatusOK, "successful-unlink", c.Param("livestreamID"))
}

func (h *Handlers) obsLinkToYouTube(c echo.Context) error {
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	strm, err := h.ls.Get(c.Request().Context(), strmID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	broadcasts, err := h.yt.GetBroadcasts(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get youtube broadcasts: %w", err)
	}
	data := struct {
		Livestream livestream.Livestream
		Broadcasts []youtube.Broadcast
	}{
		Livestream: strm,
		Broadcasts: broadcasts,
	}
	return c.Render(http.StatusOK, "set-youtube-link", data)
}

func (h *Handlers) obsLinkToYouTubeConfirm(c echo.Context) error {
	err := h.yt.EnableShowTimeForBroadcast(c.Request().Context(), c.FormValue("broadcastID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	err = h.ls.UpdateYouTubeLink(c.Request().Context(), strmID, c.FormValue("broadcastID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.Render(http.StatusCreated, "successful-link", c.Param("livestreamID"))
}

func (h *Handlers) obsUnlinkFromYouTube(c echo.Context) error {
	err := h.yt.DisableShowTimeForBroadcast(c.Request().Context(), c.Param("linkID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	err = h.ls.UpdateYouTubeLink(c.Request().Context(), strmID, "")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.Render(http.StatusOK, "successful-unlink", c.Param("livestreamID"))
}

func (h *Handlers) obsListChannels(c echo.Context) error {
	ch, err := h.mcr.ListChannels(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	data := struct {
		Channels []mcr.Channel
	}{
		Channels: ch,
	}
	return c.Render(http.StatusOK, "list-channels", data)
}

func (h *Handlers) obsNewChannel(c echo.Context) error {
	return c.Render(http.StatusOK, "new-channel", nil)
}

func (h *Handlers) obsNewChannelSubmit(c echo.Context) error {
	ch := mcr.NewChannel{}
	err := c.Bind(&ch)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	_, err = h.mcr.NewChannel(c.Request().Context(), ch)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return h.obsListChannels(c)
}
