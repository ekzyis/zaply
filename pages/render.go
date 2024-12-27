package pages

import (
	"context"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/ekzyis/zaply/env"
	"github.com/labstack/echo/v4"
)

var baseUrlContextKey = "baseUrl"
var envContextKey = "env"

func GetBaseUrl(ctx context.Context) string {
	if u, ok := ctx.Value(baseUrlContextKey).(string); ok {
		return strings.TrimRight(u, "/")
	}
	return ""
}

func GetEnv(ctx context.Context) string {
	if u, ok := ctx.Value(envContextKey).(string); ok {
		return u
	}
	return "development"
}

func OverlayHandler(c echo.Context) error {
	return render(c, http.StatusOK, Overlay())
}

func render(ctx echo.Context, statusCode int, t templ.Component) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	renderContext := context.WithValue(ctx.Request().Context(), baseUrlContextKey, env.PublicUrl)
	renderContext = context.WithValue(renderContext, envContextKey, env.Env)

	if err := t.Render(renderContext, buf); err != nil {
		return err
	}

	return ctx.HTML(statusCode, buf.String())
}
