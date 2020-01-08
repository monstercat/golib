package sse

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gin-contrib/sse"
	"github.com/gin-gonic/gin"
)

func TestClient(t *testing.T) {

	// events
	events := []sse.Event{
		{
			Event: "Heartbeat",
			Data:  "0\n1\n2",
		},
		{
			Event: "Heartbeat",
			Data:  "4\n5\n6",
		},
	}

	// Start gin server.
	g := gin.Default()
	g.GET("/notifications", func(c *gin.Context) {
		c.Stream(func(w io.Writer) bool {
			for _, e := range events {
				c.Render(-1, e)
				time.Sleep(1 * time.Second)
			}
			return true
		})
	})
	go func() {
		g.Run(":15933")
	}()

	// Start client
	errCh := make(chan error)
	evCh := make(chan Event)

	req := func() (*http.Request, error) {
		req, _ := http.NewRequest("GET", "http://localhost:15933/notifications", nil)
		req.Header.Set("Accept", "text/event-stream")
		return req, nil
	}

	Run(req, evCh, errCh, nil)

	i := 0
	for {
		select {
		case err := <-errCh:
			panic(err)
		case ev := <-evCh:
			curr := events[i]
			if curr.Event != ev.Event {
				t.Errorf("Event %d: should be %s, but got %s", i, curr.Event, ev.Event)
			}
			if curr.Data != ev.Data {
				t.Errorf("Data %d: should be %s, but got %s", i, curr.Data, ev.Data)
			}
			i++
		case <-time.After(6 * time.Second):
			panic("Timeout")
		}

		if i >= len(events) {
			return
		}
	}

}
