package lnurl

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/btcsuite/btcutil/bech32"
	"github.com/ekzyis/zaply/env"
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
	callback, err := url.JoinPath(env.PublicUrl, "/.well-known/lnurlp/", c.Param("name"), "/pay")
	if err != nil {
		return err
	}

	return c.JSON(
		http.StatusOK,
		map[string]any{
			"callback":       callback,
			"minSendable":    MIN_SENDABLE_AMOUNT,
			"maxSendable":    MAX_SENDABLE_AMOUNT,
			"metadata":       lnurlMetadata(c),
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

func lnurlMetadata(c echo.Context) string {
	s := "["
	s += fmt.Sprintf("[\"text/plain\",\"Paying %s@%s\"]", c.Param("name"), c.Request().Host)
	s += fmt.Sprintf(",[\"text/identifier\",\"%s@%s\"]", c.Param("name"), c.Request().Host)
	s += "]"
	return s
}

func lnurlError(c echo.Context, code int, err error) error {
	return c.JSON(code, map[string]any{"status": "ERROR", "error": err.Error()})
}

func Encode(base string, parts ...string) string {
	u, err := url.JoinPath(base, parts...)
	if err != nil {
		log.Fatal(err)
	}

	bech32Url, err := bech32.ConvertBits([]byte(u), 8, 5, true)
	if err != nil {
		log.Fatal(err)
	}

	lnurl, err := bech32.Encode("lnurl", bech32Url)
	if err != nil {
		log.Fatal(err)
	}

	return strings.ToUpper(lnurl)
}
