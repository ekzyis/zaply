package server

import (
	"github.com/ekzyis/zaply/env"
	"github.com/ekzyis/zaply/lightning/phoenixd"
	"github.com/ekzyis/zaply/lnurl"
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

	p := phoenixd.NewPhoenixd(
		phoenixd.WithPhoenixdURL(env.PhoenixdURL),
		phoenixd.WithPhoenixdLimitedAccessToken(env.PhoenixdLimitedAccessToken),
	)

	lnurl.Router(s.Echo, p)

	return s
}

func (s *Server) Start(address string) {
	s.Logger.Fatal(s.Echo.Start(address))
}
