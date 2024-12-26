package lnurl

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ekzyis/zaply/lightning"
	"github.com/labstack/echo/v4"
)

var (
	MIN_SENDABLE_AMOUNT = 1000            // 1 sat
	MAX_SENDABLE_AMOUNT = 100_000_000_000 // 100m sat
	MAX_COMMENT_LENGTH  = 128
)

func Router(e *echo.Echo, ln lightning.Lightning) {
	e.GET("/.well-known/lnurlp/:name", payRequest)
	e.GET("/.well-known/lnurlp/:name/pay", pay(ln))
}

func payRequest(c echo.Context) error {
	name := c.Param("name")
	return c.JSON(
		http.StatusOK,
		map[string]any{
			"callback":       fmt.Sprintf("%s/.well-known/lnurlp/%s/pay", c.Request().Host, name),
			"minSendable":    MIN_SENDABLE_AMOUNT,
			"maxSendable":    MAX_SENDABLE_AMOUNT,
			"metadata":       fmt.Sprintf("[[\"text/plain\",\"paying %s\"]]", name),
			"tag":            "payRequest",
			"commentAllowed": MAX_COMMENT_LENGTH,
		},
	)
}

func pay(ln lightning.Lightning) echo.HandlerFunc {
	return func(c echo.Context) error {
		qAmount := c.QueryParam("amount")
		if qAmount == "" {
			return lnurlError(c, http.StatusBadRequest, errors.New("amount required"))
		}
		msats, err := strconv.ParseInt(qAmount, 10, 64)
		if err != nil {
			c.Logger().Error(err)
			return lnurlError(c, http.StatusBadRequest, errors.New("invalid amount"))
		}
		if msats < 1000 {
			return lnurlError(c, http.StatusBadRequest, errors.New("amount must be at least 1000 msats"))
		}

		comment := c.QueryParam("comment")

		pr, err := ln.CreateInvoice(msats, comment)
		if err != nil {
			c.Logger().Error(err)
			return lnurlError(c, http.StatusInternalServerError, errors.New("failed to create invoice"))
		}

		return c.JSON(
			http.StatusOK,
			map[string]any{
				"pr":     pr,
				"routes": []string{},
			},
		)
	}
}

func lnurlError(c echo.Context, code int, err error) error {
	return c.JSON(code, map[string]any{"status": "ERROR", "error": err.Error()})
}
