package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/a-h/templ"
	"github.com/ekzyis/zaply/components"
	"github.com/ekzyis/zaply/lightning"
	"github.com/labstack/echo/v4"
)

type Event struct {
	Id    []byte
	Event []byte
	Data  []byte
}

func (ev *Event) MarshalTo(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "event: %s\n", ev.Event); err != nil {
		return err
	}

	for _, line := range bytes.Split(ev.Data, []byte("\n")) {
		if _, err := fmt.Fprintf(w, "data: %s\n", line); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(w, "id: %s\n", ev.Id); err != nil {
		return err
	}

	if _, err := fmt.Fprint(w, "\n"); err != nil {
		return err
	}

	return nil
}

func sseHandler(invSrc chan *lightning.Invoice) echo.HandlerFunc {
	return func(c echo.Context) error {
		w := c.Response()
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		// disable nginx buffering
		w.Header().Set("X-Accel-Buffering", "no")

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-c.Request().Context().Done():
				return nil
			case <-ticker.C:
				event := Event{
					Event: []byte("message"),
					Data:  []byte("keepalive"),
				}

				if err := event.MarshalTo(w); err != nil {
					return err
				}
			case inv := <-invSrc:
				buf := templ.GetBuffer()
				defer templ.ReleaseBuffer(buf)

				if err := components.Zap(inv).Render(c.Request().Context(), buf); err != nil {
					return err
				}

				event := Event{
					Id:    []byte(inv.PaymentHash),
					Event: []byte("zap"),
					Data:  buf.Bytes(),
				}

				log.Printf("sending zap event: %s", inv.PaymentHash)

				if err := event.MarshalTo(w); err != nil {
					return err
				}
			}
			w.Flush()
		}
	}
}
