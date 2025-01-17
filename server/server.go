package server

import (
	"fmt"
	"log"
	"net/url"

	"github.com/ekzyis/zaply/env"
	"github.com/ekzyis/zaply/lightning/phoenixd"
	"github.com/ekzyis/zaply/lnurl"
	"github.com/ekzyis/zaply/pages"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	*echo.Echo
}

func NewServer() *Server {
	s := &Server{
		Echo: echo.New(),
	}

	s.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format:           "${time_custom} ${method} ${uri} ${status}\n",
		CustomTimeFormat: "2006-01-02 15:04:05.00000-0700",
	}))

	webhookPath := "/overlay/webhook"
	webhookUrl, err := url.JoinPath(env.PublicUrl, webhookPath)
	if err != nil {
		log.Fatal(err)
	}

	p := phoenixd.NewPhoenixd(
		phoenixd.WithPhoenixdURL(env.PhoenixdURL),
		phoenixd.WithPhoenixdLimitedAccessToken(env.PhoenixdLimitedAccessToken),
		phoenixd.WithPhoenixdWebhookUrl(webhookUrl),
	)

	s.POST(webhookPath, p.WebhookHandler)

	lnurl.Router(s.Echo, p)

	s.Static("/", "public/")

	s.GET("/overlay", pages.OverlayHandler(
		lnurl.Encode(fmt.Sprintf("%s/.well-known/lnurlp/%s", env.PublicUrl, "SNL")),
	))
	s.GET("/overlay/sse", sseHandler(p.IncomingPayments()))

	return s
}

func (s *Server) Start(address string) {
	s.Logger.Fatal(s.Echo.Start(address))
}
