package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handlers) googleCallback(c echo.Context) error {
	// Check state cookie to make sure there isn't any CSRF biz
	state, _ := c.Cookie(h.stateCookieName)

	if c.FormValue("state") != state.Value {
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	code := c.FormValue("code")
	if code == "" {
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	tok, err := h.auth.NewToken(c.Request().Context(), code)
	if err != nil {
		err = fmt.Errorf("failed to get token: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	err = h.auth.StoreToken("me", tok)
	if err != nil {
		err = fmt.Errorf("failed to store token: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.String(http.StatusOK, "login successful!")
}
