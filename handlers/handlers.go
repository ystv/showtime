package handlers

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/ystv/showtime/auth"
	"github.com/ystv/showtime/playout"
)

type Handlers struct {
	auth            *auth.Auther
	play            *playout.Playouter
	mux             *echo.Echo
	stateCookieName string
}

func New(db *sqlx.DB, auth *auth.Auther) *Handlers {
	return &Handlers{
		auth:            auth,
		play:            playout.New(db),
		mux:             echo.New(),
		stateCookieName: "state-token",
	}
}

func (h *Handlers) Start() {
	h.mux.GET("/", h.index)
	h.mux.POST("/api/playouts", h.newPlayout)
	h.mux.GET("/api/playouts", h.listPlayouts)
	h.mux.GET("/api/streams", h.listStreams)
	h.mux.GET("/oauth/google/login", h.googleLogin)
	h.mux.GET("/oauth/google/callback", h.googleCallback)

	h.mux.Pre(middleware.RemoveTrailingSlash())
	h.mux.Use(middleware.Logger())
	h.mux.Use(middleware.Recover())
	h.mux.HideBanner = true

	h.mux.Logger.Fatal(h.mux.Start(":8080"))
}

func (h *Handlers) index(c echo.Context) error {
	return c.String(http.StatusOK, "it's show time!")
}
