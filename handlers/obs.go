package handlers

import (
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

func (h *Handlers) obsLinkToYouTubeSelectAccount(c echo.Context) error {
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
	}{
		Livestream: strm,
		Channels:   ch,
	}
	return c.Render(http.StatusOK, "set-youtube-link-account", data)
}

func (h *Handlers) obsLinkToYouTubeSelectBroadcast(c echo.Context) error {
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

func (h *Handlers) obsLinkToYouTubeConfirm(c echo.Context) error {
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
		return echo.NewHTTPError(http.StatusNotFound, err)
	}

	err = yt.NewExistingBroadcast(ctx, newExistingBroadcast.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	err = h.ls.UpdateYouTubeLink(ctx, strmID, newExistingBroadcast.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.Render(http.StatusCreated, "successful-link", strmID)
}

func (h *Handlers) obsUnlinkFromYouTube(c echo.Context) error {
	ctx := c.Request().Context()
	linkID := c.Param("linkID")
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	err = h.yt.DeleteExistingBroadcast(ctx, linkID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	err = h.ls.UpdateYouTubeLink(ctx, strmID, "")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.Render(http.StatusOK, "successful-unlink", strmID)
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
