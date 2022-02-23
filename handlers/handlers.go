package handlers

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/ystv/showtime/auth"
	"github.com/ystv/showtime/playout"
	"github.com/ystv/showtime/youtube"
)

type Handlers struct {
	auth            *auth.Auther
	play            *playout.Playouter
	yt              *youtube.YouTuber
	mux             *echo.Echo
	stateCookieName string
}

var domainName = "dev.ystv.co.uk"

var corsConfig middleware.CORSConfig = middleware.CORSConfig{
	AllowCredentials: true,
	Skipper:          middleware.DefaultSkipper,
	AllowOrigins: []string{
		"http://creator." + domainName,
		"https://creator." + domainName,
		"http://my." + domainName,
		"https://my." + domainName,
		"http://local." + domainName + ":3000",
		"https://local." + domainName + ":3000",
		"http://" + domainName,
		"https://" + domainName},
	AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAccessControlAllowCredentials, echo.HeaderAccessControlAllowOrigin},
	AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
}

func New(db *sqlx.DB, auth *auth.Auther, t *Templater) *Handlers {
	yt, _ := youtube.New(db, auth)

	e := echo.New()
	e.Renderer = t

	return &Handlers{
		auth:            auth,
		play:            playout.New("rtmp://example.com/app", db, yt),
		yt:              yt,
		mux:             e,
		stateCookieName: "state-token",
	}
}

func (h *Handlers) Start() {
	h.mux.GET("/", h.obsListPlayouts)
	h.mux.GET("/playouts/:playoutID", h.obsGetPlayout)
	h.mux.GET("/playouts/:playoutID/manage", h.obsManagePlayout)
	h.mux.GET("/playouts/:playoutID/link/youtube", h.obsLinkToYouTube)
	h.mux.POST("/playouts/:playoutID/link/youtube/confirm", h.obsLinkToYouTubeConfirm)
	h.mux.POST("/api/playouts", h.newPlayout)
	h.mux.PUT("/api/playouts", h.updatePlayout)
	h.mux.GET("/api/playouts", h.listPlayouts)
	h.mux.POST("/api/playouts/:playoutID/refresh-key", h.refreshStreamKey)
	h.mux.POST("/api/playouts/:playoutID/link/youtube/:broadcastID", h.enableYouTube)
	h.mux.POST("/api/playouts/:playoutID/unlink/youtube/:broadcastID", h.disableYouTube)
	h.mux.GET("/api/youtube/broadcasts", h.listYouTubeBroadcasts)
	h.mux.POST("/api/nginx/hook", h.hookStreamStart)
	h.mux.GET("/oauth/google/login", h.loginGoogle)
	h.mux.GET("/oauth/google/callback", h.callbackGoogle)

	h.mux.Pre(middleware.RemoveTrailingSlash())
	h.mux.Use(middleware.Logger())
	h.mux.Use(middleware.Recover())
	h.mux.Use(middleware.CORSWithConfig(corsConfig))
	h.mux.HideBanner = true

	h.mux.Logger.Fatal(h.mux.Start(":8080"))
}

type Templater struct {
	templates *template.Template
}

func NewTemplater(fs fs.FS) (*Templater, error) {
	t, err := template.ParseFS(fs, "*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}
	return &Templater{templates: t}, nil
}

func (t *Templater) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
