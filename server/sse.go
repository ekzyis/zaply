package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/ekzyis/zaply/lightning"
	"github.com/labstack/echo/v4"
)

type Event struct {
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
				data, err := json.Marshal(inv)
				if err != nil {
					return err
				}

				event := Event{
					Event: []byte("zap"),
					Data:  data,
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
