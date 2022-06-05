package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/ystv/showtime/livestream"
)

func (h *Handlers) listYouTubeBroadcasts(c echo.Context) error {
	broadcasts, err := h.yt.ListBroadcasts(c.Request().Context())
	if err != nil {
		err = fmt.Errorf("failed to get streams: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, broadcasts)
}

func (h *Handlers) enableYouTube(c echo.Context) error {
	ctx := c.Request().Context()
	accountID, err := strconv.Atoi(c.Param("accountID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	broadcastID := c.Param("broadcastID")
	strmID, err := strconv.Atoi(c.Param("livestreamID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	yt, err := h.yt.GetYouTuber(accountID)
	if err != nil {
		err = fmt.Errorf("failed to get youtuber: %w", err)
		return echo.NewHTTPError(http.StatusNotFound, err)
	}

	err = yt.NewExistingBroadcast(ctx, broadcastID)
	if err != nil {
		err = fmt.Errorf("failed to create new existing broadcast: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	_, err = h.ls.NewLink(ctx, livestream.NewLinkParams{
		LivestreamID:    strmID,
		IntegrationType: livestream.YTExisting,
		IntegrationID:   broadcastID,
	})
	if err != nil {
		err = fmt.Errorf("failed to create new link: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.NoContent(http.StatusCreated)
}

func (h *Handlers) disableYouTube(c echo.Context) error {
	ctx := c.Request().Context()
	linkID, err := strconv.Atoi(c.Param("linkID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	link, err := h.ls.GetLink(ctx, linkID)
	if err != nil {
		err = fmt.Errorf("failed to get link: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	err = h.yt.DeleteExistingBroadcast(ctx, link.IntegrationID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	err = h.ls.DeleteLink(ctx, linkID)
	if err != nil {
		err = fmt.Errorf("failed to delete link: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.NoContent(http.StatusOK)
}
