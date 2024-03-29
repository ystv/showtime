package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

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
	upcoming := []livestream.Livestream{}
	past := []livestream.Livestream{}
	for _, strm := range strms {
		if strm.Status == "stream-ended" {
			past = append(past, strm)
		} else {
			upcoming = append(upcoming, strm)
		}
	}

	data := struct {
		Upcoming []livestream.Livestream
		Past     []livestream.Livestream
	}{
		Upcoming: upcoming,
		Past:     past,
	}

	return c.Render(http.StatusOK, "list-livestreams", data)
}

func (h *Handlers) obsGetLivestream(c echo.Context) error {
	ctx := c.Request().Context()
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	strm, err := h.ls.Get(ctx, strmID)
	if err != nil {
		return fmt.Errorf("failed to get stream: %w", err)
	}
	return c.Render(http.StatusOK, "get-livestream", strm)
}

func (h *Handlers) obsNewLivestream(c echo.Context) error {
	return c.Render(http.StatusOK, "edit-livestream", editLivestreamForm{
		Title:  "New",
		Action: "Create",
	})
}

func (h *Handlers) obsEditLivestream(c echo.Context) error {
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	strm, err := h.ls.Get(c.Request().Context(), strmID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.Render(http.StatusOK, "edit-livestream", editLivestreamForm{
		Fields: EditLivestreamFormFields{
			Title:          strm.Title,
			Description:    strm.Description,
			ScheduledStart: strm.ScheduledStart.Format("2006-01-02T15:04"),
			ScheduledEnd:   strm.ScheduledEnd.Format("2006-01-02T15:04"),
			Visibility:     strm.Visibility,
		},
		ID:     strmID,
		Title:  "Edit",
		Action: "Save",
	})
}

type (
	editLivestreamForm struct {
		Fields EditLivestreamFormFields
		ID     int
		Title  string
		Action string
		Errors []string
	}
	// EditLivestreamFormFields are fields on the form.
	EditLivestreamFormFields struct {
		Title          string `form:"title"`
		Description    string `form:"description"`
		ScheduledStart string `form:"scheduledStart"`
		ScheduledEnd   string `form:"scheduledEnd"`
		Visibility     string `form:"visibility"`
	}
)

func (h *Handlers) obsNewLivestreamSubmit(c echo.Context) error {
	form := editLivestreamForm{
		Title:  "New",
		Action: "Create",
	}

	err := c.Bind(&form.Fields)
	if err != nil {
		form.Errors = append(form.Errors, err.Error())
		return c.Render(http.StatusBadRequest, "edit-livestream", form)
	}

	if form.Fields.ScheduledStart == "" {
		form.Errors = append(form.Errors, "scheduled start is required")
	}
	if form.Fields.ScheduledEnd == "" {
		form.Errors = append(form.Errors, "scheduled end is required")
	}

	if len(form.Errors) != 0 {
		return c.Render(http.StatusBadRequest, "edit-livestream", form)
	}
	scheduledStart, err := time.Parse(time.RFC3339, form.Fields.ScheduledStart+":00Z")
	if err != nil {
		form.Errors = append(form.Errors, err.Error())
		return c.Render(http.StatusBadRequest, "edit-livestream", form)
	}
	scheduledEnd, err := time.Parse(time.RFC3339, fmt.Sprintf("%s:00Z", form.Fields.ScheduledEnd))
	if err != nil {
		form.Errors = append(form.Errors, err.Error())
		return c.Render(http.StatusBadRequest, "edit-livestream", form)
	}

	strm := livestream.EditLivestream{
		Title:          form.Fields.Title,
		Description:    form.Fields.Description,
		ScheduledStart: scheduledStart,
		ScheduledEnd:   scheduledEnd,
		Visibility:     form.Fields.Visibility,
	}
	strmID, err := h.ls.New(c.Request().Context(), strm)
	if err != nil {
		form.Errors = append(form.Errors, err.Error())
		return c.Render(http.StatusBadRequest, "edit-livestream", form)
	}
	return c.Redirect(http.StatusFound, fmt.Sprintf("/livestreams/%d", strmID))
}

func (h *Handlers) obsEditLivestreamSubmit(c echo.Context) error {
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	form := editLivestreamForm{
		ID:     strmID,
		Title:  "Edit",
		Action: "Save",
	}

	err = c.Bind(&form.Fields)
	if err != nil {
		form.Errors = append(form.Errors, err.Error())
		return c.Render(http.StatusBadRequest, "edit-livestream", form)
	}

	if form.Fields.ScheduledStart == "" {
		form.Errors = append(form.Errors, "scheduled start is required")
	}
	if form.Fields.ScheduledEnd == "" {
		form.Errors = append(form.Errors, "scheduled end is required")
	}

	if len(form.Errors) != 0 {
		return c.Render(http.StatusBadRequest, "edit-livestream", form)
	}
	scheduledStart, err := time.Parse(time.RFC3339, form.Fields.ScheduledStart+":00Z")
	if err != nil {
		form.Errors = append(form.Errors, err.Error())
		return c.Render(http.StatusBadRequest, "edit-livestream", form)
	}
	scheduledEnd, err := time.Parse(time.RFC3339, fmt.Sprintf("%s:00Z", form.Fields.ScheduledEnd))
	if err != nil {
		form.Errors = append(form.Errors, err.Error())
		return c.Render(http.StatusBadRequest, "edit-livestream", form)
	}

	strm := livestream.EditLivestream{
		Title:          form.Fields.Title,
		Description:    form.Fields.Description,
		ScheduledStart: scheduledStart,
		ScheduledEnd:   scheduledEnd,
		Visibility:     form.Fields.Visibility,
	}
	err = h.ls.Update(c.Request().Context(), strmID, strm)
	if err != nil {
		form.Errors = append(form.Errors, err.Error())
		return c.Render(http.StatusBadRequest, "edit-livestream", form)
	}
	return c.Redirect(http.StatusFound, fmt.Sprintf("/livestreams/%d", strmID))
}

func (h *Handlers) obsStartLivestream(c echo.Context) error {
	ctx := c.Request().Context()
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	strm, err := h.ls.Get(ctx, strmID)
	if err != nil {
		return fmt.Errorf("failed to get livestream: %w", err)
	}
	err = h.ls.Start(ctx, strm)
	if err != nil {
		return fmt.Errorf("failed to start livestream: %w", err)
	}

	return h.obsGetLivestream(c)
}

func (h *Handlers) obsEndLivestream(c echo.Context) error {
	ctx := c.Request().Context()
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	strm, err := h.ls.Get(ctx, strmID)
	if err != nil {
		return fmt.Errorf("failed to get livestream: %w", err)
	}
	err = h.ls.End(ctx, strm)
	if err != nil {
		return fmt.Errorf("failed to end livestream: %w", err)
	}

	return h.obsListLivestreams(c)
}

func (h *Handlers) obsManageLivestream(c echo.Context) error {
	ctx := c.Request().Context()
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	strm, err := h.ls.Get(ctx, strmID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	links, err := h.ls.ListLinks(ctx, strmID)
	if err != nil {
		return fmt.Errorf("failed to get stream links: %w", err)
	}

	data := struct {
		Livestream livestream.Livestream
		Links      []livestream.Link
	}{
		Livestream: strm,
		Links:      links,
	}
	return c.Render(http.StatusOK, "manage-livestream", data)
}

func (h *Handlers) obsDeleteLivestream(c echo.Context) error {
	ctx := c.Request().Context()
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	strm, err := h.ls.Get(ctx, strmID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	links, err := h.ls.ListLinks(ctx, strmID)
	if err != nil {
		return fmt.Errorf("failed to get stream links: %w", err)
	}

	data := struct {
		Livestream livestream.Livestream
		TotalLinks int
	}{
		Livestream: strm,
		TotalLinks: len(links),
	}
	return c.Render(http.StatusOK, "delete-livestream", data)
}

func (h *Handlers) obsDeleteLivestreamSubmit(c echo.Context) error {
	ctx := c.Request().Context()
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	strm, err := h.ls.Get(ctx, strmID)
	if err != nil {
		err = fmt.Errorf("failed to get livestream: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	err = h.ls.Delete(ctx, strm)
	if err != nil {
		err = fmt.Errorf("failed to delete livestream: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.Render(http.StatusOK, "successful-livestream-delete", nil)
}

func (h *Handlers) obsLink(c echo.Context) error {
	ctx := c.Request().Context()
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	strm, err := h.ls.Get(ctx, strmID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.Render(http.StatusOK, "new-link", strm)
}

func (h *Handlers) obsUnlink(c echo.Context) error {
	ctx := c.Request().Context()
	linkID, err := strconv.Atoi(c.Param("linkID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	link, err := h.ls.GetLink(ctx, linkID)
	if err != nil {
		return fmt.Errorf("failed to get link: %w", err)
	}

	err = h.ls.DeleteLink(ctx, link)
	if err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}

	return c.Render(http.StatusOK, "successful-unlink", strmID)
}

func (h *Handlers) obsLinkToMCR(c echo.Context) error {
	ctx := c.Request().Context()
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	strm, err := h.ls.Get(ctx, strmID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	ch, err := h.mcr.ListChannels(ctx)
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
	ctx := c.Request().Context()
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	strm, err := h.ls.Get(ctx, strmID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	res := struct {
		ChannelID int `form:"channelID"`
	}{}
	err = c.Bind(&res)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	po := mcr.EditPlayout{
		ChannelID:      res.ChannelID,
		SrcURI:         h.conf.IngestAddress + "/" + strm.StreamKey,
		Title:          strm.Title,
		Description:    strm.Description,
		ScheduledStart: strm.ScheduledStart,
		ScheduledEnd:   strm.ScheduledEnd,
		Visibility:     strm.Visibility,
	}
	playoutID, err := h.mcr.NewPlayout(ctx, po)
	if err != nil {
		err = fmt.Errorf("failed to create new playout: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	_, err = h.ls.NewLink(ctx, livestream.NewLinkParams{
		LivestreamID:    strmID,
		IntegrationType: livestream.LinkMCR,
		IntegrationID:   strconv.Itoa(playoutID),
	})
	if err != nil {
		err = fmt.Errorf("failed to create new link: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.Render(http.StatusCreated, "successful-link", strmID)
}

func (h *Handlers) obsLinkToYouTube(c echo.Context) error {
	ctx := c.Request().Context()
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	strm, err := h.ls.Get(ctx, strmID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	ch, err := h.yt.About(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	data := struct {
		Livestream livestream.Livestream
		Channels   []youtube.ChannelInfo
		Action     string
	}{
		Livestream: strm,
		Channels:   ch,
		Action:     "create",
	}
	return c.Render(http.StatusOK, "set-youtube-link-account", data)
}

func (h *Handlers) obsLinkToYouTubeConfirm(c echo.Context) error {
	ctx := c.Request().Context()

	accountID, err := strconv.Atoi(c.FormValue("accountID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	strm, err := h.ls.Get(ctx, strmID)
	if err != nil {
		err = fmt.Errorf("failed to get livestream: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	yt, err := h.yt.GetYouTuber(accountID)
	if err != nil {
		err = fmt.Errorf("failed to get youtuber: %w", err)
		return echo.NewHTTPError(http.StatusNotFound, err)
	}

	b, err := yt.NewBroadcast(ctx, youtube.EditBroadcast{
		Title:          strm.Title,
		Description:    strm.Description,
		ScheduledStart: strm.ScheduledStart,
		ScheduledEnd:   strm.ScheduledEnd,
		Visibility:     strm.Visibility,
	})
	if err != nil {
		err = fmt.Errorf("failed to create new broadcast: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	_, err = h.ls.NewLink(ctx, livestream.NewLinkParams{
		LivestreamID:    strmID,
		IntegrationType: livestream.LinkYTNew,
		IntegrationID:   b.ID,
	})
	if err != nil {
		err = fmt.Errorf("failed to create new link: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.Render(http.StatusCreated, "successful-link", strmID)
}

func (h *Handlers) obsLinkToYouTubeExistingSelectAccount(c echo.Context) error {
	ctx := c.Request().Context()
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	strm, err := h.ls.Get(ctx, strmID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	ch, err := h.yt.About(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	data := struct {
		Livestream livestream.Livestream
		Channels   []youtube.ChannelInfo
		Action     string
	}{
		Livestream: strm,
		Channels:   ch,
		Action:     "select-broadcast",
	}
	return c.Render(http.StatusOK, "set-youtube-link-account", data)
}

func (h *Handlers) obsLinkToYouTubeExistingSelectBroadcast(c echo.Context) error {
	ctx := c.Request().Context()
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	strm, err := h.ls.Get(ctx, strmID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	accountID, err := strconv.Atoi(c.FormValue("accountID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	yt, err := h.yt.GetYouTuber(accountID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err)
	}

	broadcasts, err := yt.ListBroadcasts(ctx)
	if err != nil {
		return fmt.Errorf("failed to get youtube broadcasts: %w", err)
	}
	data := struct {
		AccountID  int
		Livestream livestream.Livestream
		Broadcasts []youtube.Broadcast
	}{
		AccountID:  accountID,
		Livestream: strm,
		Broadcasts: broadcasts,
	}
	return c.Render(http.StatusOK, "set-youtube-link-broadcast", data)
}

func (h *Handlers) obsLinkToYouTubeExistingConfirm(c echo.Context) error {
	ctx := c.Request().Context()

	newExistingBroadcast := struct {
		ID        string `form:"broadcastID"`
		AccountID string `form:"accountID"`
	}{}
	err := c.Bind(&newExistingBroadcast)
	if err != nil {
		err = fmt.Errorf("failed to bind form response: %w", err)
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	accountID, err := strconv.Atoi(newExistingBroadcast.AccountID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	yt, err := h.yt.GetYouTuber(accountID)
	if err != nil {
		err = fmt.Errorf("failed to get youtuber: %w", err)
		return echo.NewHTTPError(http.StatusNotFound, err)
	}

	err = yt.NewExistingBroadcast(ctx, newExistingBroadcast.ID)
	if err != nil {
		err = fmt.Errorf("failed to create new existing broadcast: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	_, err = h.ls.NewLink(ctx, livestream.NewLinkParams{
		LivestreamID:    strmID,
		IntegrationType: livestream.LinkYTExisting,
		IntegrationID:   newExistingBroadcast.ID,
	})
	if err != nil {
		err = fmt.Errorf("failed to create new link: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.Render(http.StatusCreated, "successful-link", strmID)
}

func (h *Handlers) obsLinkToRTMP(c echo.Context) error {
	ctx := c.Request().Context()
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	strm, err := h.ls.Get(ctx, strmID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	data := struct {
		Livestream livestream.Livestream
	}{
		Livestream: strm,
	}
	return c.Render(http.StatusOK, "set-rtmp-output-link", data)
}

func (h *Handlers) obsLinkToRTMPConfirm(c echo.Context) error {
	ctx := c.Request().Context()
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	strm, err := h.ls.Get(ctx, strmID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	outputURL := c.FormValue("outputURL")

	rtmpOutput, err := h.ls.NewRTMPOutput(ctx, outputURL)
	if err != nil {
		err = fmt.Errorf("failed to create new rtmp output: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	_, err = h.ls.NewLink(ctx, livestream.NewLinkParams{
		LivestreamID:    strm.ID,
		IntegrationType: livestream.LinkRTMPOutput,
		IntegrationID:   strconv.Itoa(rtmpOutput.ID),
	})
	if err != nil {
		err = fmt.Errorf("failed to create new link: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.Render(http.StatusCreated, "successful-link", strmID)
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
	return c.Render(http.StatusOK, "edit-channel", editChannelForm{
		Fields: mcr.EditChannel{},
		Title:  "New",
		Action: "Create",
	})
}

type (
	editChannelForm struct {
		ID     int
		Fields mcr.EditChannel
		Title  string
		Action string
		Errors []string
	}
)

func (h *Handlers) obsNewChannelSubmit(c echo.Context) error {
	form := editChannelForm{
		Fields: mcr.EditChannel{},
		Title:  "New",
		Action: "Create",
	}

	err := c.Bind(&form.Fields)
	if err != nil {
		form.Errors = append(form.Errors, err.Error())
		return c.Render(http.StatusBadRequest, "edit-channel", form)
	}

	chID, err := h.mcr.NewChannel(c.Request().Context(), form.Fields)
	if err != nil {
		form.Errors = append(form.Errors, err.Error())
		return c.Render(http.StatusBadRequest, "edit-channel", form)
	}
	return c.Redirect(http.StatusFound, fmt.Sprintf("/channels/%d", chID))
}

func (h *Handlers) obsGetChannel(c echo.Context) error {
	ctx := c.Request().Context()
	channelID, err := strconv.Atoi(c.Param("channelID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	ch, err := h.mcr.GetChannel(ctx, channelID)
	if err != nil {
		err = fmt.Errorf("failed to get channel: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	po, err := h.mcr.GetPlayoutsForChannel(ctx, ch)
	if err != nil {
		return fmt.Errorf("failed to get playuts: %w", err)
	}
	data := struct {
		Channel  mcr.Channel
		Playouts []mcr.Playout
	}{
		Channel:  ch,
		Playouts: po,
	}
	return c.Render(http.StatusOK, "get-channel", data)
}

func (h *Handlers) obsEditChannel(c echo.Context) error {
	ctx := c.Request().Context()
	channelID, err := strconv.Atoi(c.Param("channelID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	ch, err := h.mcr.GetChannel(ctx, channelID)
	if err != nil {
		err = fmt.Errorf("failed to get channel: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.Render(http.StatusOK, "edit-channel", editChannelForm{
		Fields: mcr.EditChannel{
			Title:   ch.Title,
			URLName: ch.URLName,
		},
		ID:     ch.ID,
		Title:  "Edit",
		Action: "Save",
	})
}

func (h *Handlers) obsEditChannelSubmit(c echo.Context) error {
	ctx := c.Request().Context()
	channelID, err := strconv.Atoi(c.Param("channelID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	ch, err := h.mcr.GetChannel(ctx, channelID)
	if err != nil {
		err = fmt.Errorf("failed to get channel: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	form := editChannelForm{
		Fields: mcr.EditChannel{},
		ID:     ch.ID,
		Title:  "Edit",
		Action: "Save",
	}

	err = c.Bind(&form.Fields)
	if err != nil {
		form.Errors = append(form.Errors, err.Error())
		return c.Render(http.StatusBadRequest, "edit-channel", form)
	}

	err = h.mcr.UpdateChannel(ctx, ch.ID, form.Fields)
	if err != nil {
		form.Errors = append(form.Errors, err.Error())
		return c.Render(http.StatusBadRequest, "edit-channel", form)
	}
	return c.Redirect(http.StatusFound, fmt.Sprintf("/channels/%d", ch.ID))
}

func (h *Handlers) obsSetChannelOnAir(c echo.Context) error {
	ctx := c.Request().Context()
	channelID, err := strconv.Atoi(c.Param("channelID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	ch, err := h.mcr.GetChannel(ctx, channelID)
	if err != nil {
		err = fmt.Errorf("failed to get channel: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	if ch.Status == "on-air" {
		err = errors.New("channel already on-air")
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	err = h.mcr.SetChannelOnAir(ctx, ch)
	if err != nil {
		err = fmt.Errorf("failed to set channel on-air: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("/channels/%d", ch.ID))
}

func (h *Handlers) obsSetChannelOffAir(c echo.Context) error {
	ctx := c.Request().Context()
	channelID, err := strconv.Atoi(c.Param("channelID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	ch, err := h.mcr.GetChannel(ctx, channelID)
	if err != nil {
		err = fmt.Errorf("failed to get channel: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	if ch.Status == "off-air" {
		err = errors.New("channel already off-air")
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	err = h.mcr.SetChannelOffAir(ctx, ch)
	if err != nil {
		err = fmt.Errorf("failed to set channel off-air: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("/channels/%d", ch.ID))
}

func (h *Handlers) obsArchiveChannel(c echo.Context) error {
	ctx := c.Request().Context()
	channelID, err := strconv.Atoi(c.Param("channelID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	ch, err := h.mcr.GetChannel(ctx, channelID)
	if err != nil {
		err = fmt.Errorf("failed to get channel: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	data := struct {
		Channel mcr.Channel
		Title   string
		Status  string
	}{
		Channel: ch,
		Title:   "Archive",
		Status:  "archive",
	}

	return c.Render(http.StatusFound, "set-channel-status", data)
}

func (h *Handlers) obsArchiveChannelConfirm(c echo.Context) error {
	ctx := c.Request().Context()
	channelID, err := strconv.Atoi(c.Param("channelID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	ch, err := h.mcr.GetChannel(ctx, channelID)
	if err != nil {
		err = fmt.Errorf("failed to get channel: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	err = h.mcr.ArchiveChannel(ctx, ch)
	if err != nil {
		err = fmt.Errorf("failed to archive channel: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.Redirect(http.StatusFound, "/channels")
}

func (h *Handlers) obsUnarchiveChannel(c echo.Context) error {
	ctx := c.Request().Context()
	channelID, err := strconv.Atoi(c.Param("channelID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	ch, err := h.mcr.GetChannel(ctx, channelID)
	if err != nil {
		err = fmt.Errorf("failed to get channel: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	data := struct {
		Channel mcr.Channel
		Title   string
		Status  string
	}{
		Channel: ch,
		Title:   "Un-archive",
		Status:  "unarchive",
	}

	return c.Render(http.StatusFound, "set-channel-status", data)
}

func (h *Handlers) obsUnarchiveChannelConfirm(c echo.Context) error {
	ctx := c.Request().Context()
	channelID, err := strconv.Atoi(c.Param("channelID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	ch, err := h.mcr.GetChannel(ctx, channelID)
	if err != nil {
		err = fmt.Errorf("failed to get channel: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	err = h.mcr.UnarchiveChannel(ctx, ch)
	if err != nil {
		err = fmt.Errorf("failed to set channel unarchived: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("/channels/%d", ch.ID))
}

func (h *Handlers) obsDeleteChannel(c echo.Context) error {
	ctx := c.Request().Context()
	channelID, err := strconv.Atoi(c.Param("channelID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	ch, err := h.mcr.GetChannel(ctx, channelID)
	if err != nil {
		err = fmt.Errorf("failed to get channel: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	data := struct {
		Channel mcr.Channel
		Title   string
		Status  string
	}{
		Channel: ch,
		Title:   "Delete",
		Status:  "delete",
	}

	return c.Render(http.StatusFound, "set-channel-status", data)
}

func (h *Handlers) obsDeleteChannelConfirm(c echo.Context) error {
	ctx := c.Request().Context()
	channelID, err := strconv.Atoi(c.Param("channelID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	ch, err := h.mcr.GetChannel(ctx, channelID)
	if err != nil {
		err = fmt.Errorf("failed to get channel: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	err = h.mcr.DeleteChannel(ctx, ch)
	if err != nil {
		err = fmt.Errorf("failed to delete channel: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.Redirect(http.StatusFound, "/channels")
}

type integrations struct {
	YouTube []youtube.ChannelInfo
}

type listIntegrationsResponse struct {
	Integrations integrations
}

func (h *Handlers) obsListIntegrations(c echo.Context) error {
	info, err := h.yt.About(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	data := listIntegrationsResponse{
		Integrations: integrations{
			YouTube: info,
		},
	}
	return c.Render(http.StatusOK, "list-integrations", data)
}

func (h *Handlers) obsDeleteYouTubeIntegration(c echo.Context) error {
	ctx := c.Request().Context()
	accountID, err := strconv.Atoi(c.Param("accountID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	yt, err := h.yt.GetYouTuber(accountID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err)
	}
	info, err := yt.About(ctx)
	if err != nil {
		err = fmt.Errorf("failed to get about info: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	total, err := yt.GetTotalLinkedBroadcasts(ctx)
	if err != nil {
		err = fmt.Errorf("failed to get total linked broadcasts: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	data := struct {
		About           []youtube.ChannelInfo
		TotalBroadcasts int
	}{
		info,
		total,
	}
	return c.Render(http.StatusOK, "delete-integration", data)
}

func (h *Handlers) obsDeleteYouTubeIntegrationConfirm(c echo.Context) error {
	ctx := c.Request().Context()
	accountID, err := strconv.Atoi(c.Param("accountID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	yt, err := h.yt.GetYouTuber(accountID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err)
	}

	b, err := yt.ListShowTimedBroadcasts(ctx)
	if err != nil {
		err = fmt.Errorf("failed to list showtimed broadcasts: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	for _, broadcastID := range b {
		err = yt.DeleteExistingBroadcast(ctx, broadcastID)
		if err != nil {
			err = fmt.Errorf("failed to unlink broadcast: %w", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
	}

	err = h.yt.DeleteAccount(ctx, accountID)
	if err != nil {
		err = fmt.Errorf("failed to delete account: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.Render(http.StatusOK, "successful-unintegration", nil)
}
