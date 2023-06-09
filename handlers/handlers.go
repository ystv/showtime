package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/ystv/showtime/auth"
	"github.com/ystv/showtime/livestream"
	"github.com/ystv/showtime/mcr"
	"github.com/ystv/showtime/youtube"
)

type (
	// Handlers is a HTTP server.
	Handlers struct {
		conf      *Config
		jwtConfig middleware.JWTConfig
		auth      *auth.Auther
		mcr       *mcr.MCR
		ls        *livestream.Livestreamer
		yt        *youtube.YouTube
		mux       *echo.Echo
	}

	// Config configures the HTTP server.
	Config struct {
		Debug           bool
		StateCookieName string
		DomainName      string
		IngestAddress   string
		JWTSigningKey   string
	}

	// JWTClaims represents an identifiable JWT
	JWTClaims struct {
		UserID      int          `json:"id"`
		Permissions []Permission `json:"perms"`
		jwt.StandardClaims
	}
	// Permission represents the permissions that a user has
	Permission struct {
		Name string `json:"name"`
	}
)

// New creates a new handler instance.
func New(conf *Config, auth *auth.Auther, ls *livestream.Livestreamer, mcr *mcr.MCR, yt *youtube.YouTube, t *Templater) *Handlers {
	e := echo.New()
	e.Renderer = t
	e.Debug = conf.Debug

	return &Handlers{
		conf: conf,
		jwtConfig: middleware.JWTConfig{
			Claims:     &JWTClaims{},
			SigningKey: []byte(conf.JWTSigningKey),
		},
		auth: auth,
		ls:   ls,
		mcr:  mcr,
		yt:   yt,
		mux:  e,
	}
}

// Start sets up a HTTP server listening.
func (h *Handlers) Start() {
	internal := h.mux.Group("")
	{
		// Basic UI endpoints
		internal.GET("/", h.obsHome)
		internal.GET("/livestreams", h.obsListLivestreams)
		internal.GET("/livestreams/new", h.obsNewLivestream)
		internal.POST("/livestreams/new", h.obsNewLivestreamSubmit)
		strm := internal.Group("/livestreams/:livestreamID")
		{
			strm.GET("", h.obsGetLivestream)
			strm.GET("/start", h.obsStartLivestream)
			strm.GET("/end", h.obsEndLivestream)
			strm.GET("/edit", h.obsEditLivestream)
			strm.POST("/edit", h.obsEditLivestreamSubmit)
			strm.GET("/manage", h.obsManageLivestream)
			strm.GET("/delete", h.obsDeleteLivestream)
			strm.POST("/delete", h.obsDeleteLivestreamSubmit)
			strm.GET("/link", h.obsLink)
			strm.GET("/unlink/:linkID", h.obsUnlink)
			strm.GET("/link/mcr", h.obsLinkToMCR)
			strm.POST("/link/mcr/confirm", h.obsLinkToMCRConfirm)
			strm.GET("/link/youtube", h.obsLinkToYouTube)
			strm.POST("/link/youtube", h.obsLinkToYouTubeConfirm)
			strm.GET("/link/youtube-existing", h.obsLinkToYouTubeExistingSelectAccount)
			strm.POST("/link/youtube-existing", h.obsLinkToYouTubeExistingSelectBroadcast)
			strm.POST("/link/youtube-existing/confirm", h.obsLinkToYouTubeExistingConfirm)
			strm.GET("/link/rtmp", h.obsLinkToRTMP)
			strm.POST("/link/rtmp", h.obsLinkToRTMPConfirm)
		}
		internal.GET("/channels", h.obsListChannels)
		internal.GET("/channels/new", h.obsNewChannel)
		internal.POST("/channels/new", h.obsNewChannelSubmit)
		ch := internal.Group("/channels/:channelID")
		{
			ch.GET("", h.obsGetChannel)
			ch.GET("/edit", h.obsEditChannel)
			ch.POST("/edit", h.obsEditChannelSubmit)
			ch.POST("/on-air", h.obsSetChannelOnAir)
			ch.POST("/off-air", h.obsSetChannelOffAir)
			ch.GET("/archive", h.obsArchiveChannel)
			ch.POST("/archive", h.obsArchiveChannelConfirm)
			ch.GET("/un-archive", h.obsUnarchiveChannel)
			ch.POST("/un-archive", h.obsUnarchiveChannelConfirm)
			ch.GET("/delete", h.obsDeleteChannel)
			ch.POST("/delete", h.obsDeleteChannelConfirm)
		}
		internal.GET("/mixers/new", h.obsNewMixer)
		internal.GET("/mixers/new/obs", h.obsNewMixerOBS)
		internal.POST("/mixers/new/obs", h.obsNewMixerOBSSubmit)
		m := internal.Group("/mixers/:mixerID")
		{
			m.GET("/edit/obs", h.obsEditMixerOBS)
			m.POST("/edit/obs", h.obsEditMixerOBSSubmit)
		}

		internal.GET("/integrations", h.obsListIntegrations)
		internal.GET("/integrations/unlink/youtube/:accountID", h.obsDeleteYouTubeIntegration)
		internal.POST("/integrations/unlink/youtube/:accountID", h.obsDeleteYouTubeIntegrationConfirm)

		// API endpoints
		api := internal.Group("/api")
		if !h.conf.Debug {
			api.Use(middleware.JWTWithConfig(h.jwtConfig))
		}
		{
			api.POST("/livestreams", h.newLivestream)
			api.PUT("/livestreams", h.updateLivestream)
			api.GET("/livestreams", h.listLivestreams)
			api.GET("/livestreams/:livestreamID/events", h.getLivestreamEvents)
			api.POST("/livestreams/:livestreamID/refresh-key", h.refreshStreamKey)
			api.POST("/livestreams/:livestreamID/link/youtube/:broadcastID", h.enableYouTube)
			api.POST("/livestreams/:livestreamID/unlink/youtube/:broadcastID", h.disableYouTube)
			api.GET("/youtube/broadcasts", h.listYouTubeBroadcasts)
		}
	}

	// Endpoints that skip authentication
	h.mux.GET("/api/health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	h.mux.GET("/api/version", func(c echo.Context) error {
		info, ok := debug.ReadBuildInfo()
		if !ok {
			return c.NoContent(http.StatusNoContent)
		}
		return c.JSON(http.StatusOK, info.Settings)
	})
	h.mux.POST("/api/hooks/nginx/on_publish", h.hookStreamStart)
	h.mux.POST("/api/hooks/nginx/on_publish_done", h.hookStreamDone)
	h.mux.GET("/oauth/google/login", h.loginGoogle)
	h.mux.GET("/oauth/google/callback", h.callbackGoogle)
	h.mux.Static("/assets", "assets")

	corsConfig := middleware.CORSConfig{
		AllowCredentials: true,
		Skipper:          middleware.DefaultSkipper,
		AllowOrigins: []string{
			"http://creator." + h.conf.DomainName,
			"https://creator." + h.conf.DomainName,
			"http://my." + h.conf.DomainName,
			"https://my." + h.conf.DomainName,
			"http://local." + h.conf.DomainName + ":3000",
			"https://local." + h.conf.DomainName + ":3000",
			"http://" + h.conf.DomainName,
			"https://" + h.conf.DomainName},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAccessControlAllowCredentials, echo.HeaderAccessControlAllowOrigin},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}

	h.mux.Pre(middleware.RemoveTrailingSlash())
	h.mux.Use(middleware.Logger())
	h.mux.Use(middleware.Recover())
	h.mux.Use(middleware.CORSWithConfig(corsConfig))
	h.mux.HideBanner = true
	h.mux.HTTPErrorHandler = h.handleError

	h.mux.Logger.Fatal(h.mux.Start(":8080"))
}

func (h *Handlers) handleError(err error, c echo.Context) {
	if err == nil {
		return
	}
	isJSON := strings.Contains(c.Request().Header.Get("Accept"), "application/json")

	// TODO(https://ystv.atlassian.net/browse/SHOW-50): this should be handled at the handler level, not here
	if errors.Is(err, sql.ErrNoRows) {
		err = echo.NewHTTPError(http.StatusNotFound, err)
	}

	var httpErr *echo.HTTPError
	if errors.As(err, &httpErr) {
		if isJSON {
			_ = c.JSON(httpErr.Code, map[string]string{"error": fmt.Sprintf("%v", httpErr.Message)})
		} else {
			_ = c.String(httpErr.Code, fmt.Sprintf("%s: %v", http.StatusText(httpErr.Code), httpErr.Message))
		}
		return
	}
	h.mux.Logger.Errorf("%s %s %s error: %v", c.Request().Method, c.Request().URL, c.Request().RemoteAddr, err)
	if isJSON {
		_ = c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error", "detail": fmt.Sprintf("%v", err)})
	} else {
		_ = c.String(http.StatusInternalServerError, fmt.Sprintf("internal server error (please check the logs for details): %v", err))
	}
}

// Templater creates webpages for UI.
type Templater struct {
	templates *template.Template
}

// NewTemplater creates a new templater instance.
func NewTemplater(fs fs.FS) (*Templater, error) {
	t, err := template.ParseFS(fs, "*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}
	return &Templater{templates: t}, nil
}

// Render takes a template and applies data to it.
func (t *Templater) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
