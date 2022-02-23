package handlers

import (
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

func New(db *sqlx.DB, auth *auth.Auther) *Handlers {
	yt, _ := youtube.New(db, auth)
	return &Handlers{
		auth:            auth,
		play:            playout.New(db),
		yt:              yt,
		mux:             echo.New(),
		stateCookieName: "state-token",
	}
}

func (h *Handlers) Start() {
	h.mux.GET("/", h.index)
	h.mux.POST("/api/playouts", h.newPlayout)
	h.mux.PUT("/api/playouts", h.updatePlayout)
	h.mux.GET("/api/playouts", h.listPlayouts)
	h.mux.POST("/api/playouts/:playoutID/link/youtube/:broadcastID", h.enableYouTube)
	h.mux.POST("/api/playouts/:playoutID/unlink/youtube/:broadcastID", h.disableYouTube)
	h.mux.GET("/api/youtube/broadcasts", h.listYouTubeBroadcasts)
	h.mux.POST("/api/nginx/hook", h.hookStreamStart)
	h.mux.GET("/oauth/google/login", h.loginGoogle)
	h.mux.GET("/oauth/google/callback", h.callbackGoogle)

	h.mux.Pre(middleware.RemoveTrailingSlash())
	h.mux.Use(middleware.Logger())
	h.mux.Use(middleware.Recover())
	h.mux.HideBanner = true

	h.mux.Logger.Fatal(h.mux.Start(":8080"))
}

func (h *Handlers) index(c echo.Context) error {
	return c.String(http.StatusOK, "it's show time!")
}
