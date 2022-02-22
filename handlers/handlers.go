package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/ystv/showtime/auth"
)

type Handlers struct {
	auth            *auth.Auther
	mux             *echo.Echo
	stateCookieName string
}

func New(auth *auth.Auther) *Handlers {
	return &Handlers{
		auth:            auth,
		mux:             echo.New(),
		stateCookieName: "state-token",
	}
}

func (h *Handlers) Start() {
	h.mux.GET("/", h.index)
	h.mux.GET("/streams", h.showStreams)
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
