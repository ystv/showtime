package handlers

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/ystv/showtime/auth"
	"github.com/ystv/showtime/playout"
	"github.com/ystv/showtime/youtube"
)

type (
	Handlers struct {
		conf      *Config
		jwtConfig middleware.JWTConfig
		auth      *auth.Auther
		play      *playout.Playouter
		yt        *youtube.YouTuber
		mux       *echo.Echo
	}

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
		PermissionID int    `json:"id"`
		Name         string `json:"name"`
	}
)

func New(conf *Config, db *sqlx.DB, auth *auth.Auther, t *Templater) *Handlers {
	yt, _ := youtube.New(db, auth)

	e := echo.New()
	e.Renderer = t
	e.Debug = conf.Debug

	return &Handlers{
		conf: conf,
		jwtConfig: middleware.JWTConfig{
			Claims:      &JWTClaims{},
			TokenLookup: "cookie:token",
			SigningKey:  []byte(conf.JWTSigningKey),
		},
		auth: auth,
		play: playout.New(conf.IngestAddress, db, yt),
		yt:   yt,
		mux:  e,
	}
}

func (h *Handlers) Start() {
	internal := h.mux.Group("")
	if !h.conf.Debug {
		internal.Use(middleware.JWTWithConfig(h.jwtConfig))
	}
	{
		internal.GET("/", h.obsListPlayouts)
		internal.GET("/playouts/:playoutID", h.obsGetPlayout)
		internal.GET("/playouts/new", h.obsNewPlayout)
		internal.POST("/playouts/new", h.obsNewPlayoutSubmit)
		internal.GET("/playouts/:playoutID/manage", h.obsManagePlayout)
		internal.GET("/playouts/:playoutID/link/youtube", h.obsLinkToYouTube)
		internal.POST("/playouts/:playoutID/link/youtube/confirm", h.obsLinkToYouTubeConfirm)
		internal.GET("/playouts/:playoutID/unlink/youtube/:broadcastID", h.obsUnlinkFromYouTube)
		internal.POST("/api/playouts", h.newPlayout)
		internal.PUT("/api/playouts", h.updatePlayout)
		internal.GET("/api/playouts", h.listPlayouts)
		internal.POST("/api/playouts/:playoutID/refresh-key", h.refreshStreamKey)
		internal.POST("/api/playouts/:playoutID/link/youtube/:broadcastID", h.enableYouTube)
		internal.POST("/api/playouts/:playoutID/unlink/youtube/:broadcastID", h.disableYouTube)
		internal.GET("/api/youtube/broadcasts", h.listYouTubeBroadcasts)
	}

	h.mux.POST("/api/nginx/hook", h.hookStreamStart)
	h.mux.GET("/oauth/google/login", h.loginGoogle)
	h.mux.GET("/oauth/google/callback", h.callbackGoogle)

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
