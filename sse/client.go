package sse

import (
	"bufio"
	"bytes"
	"net/http"
	"strings"
	"time"
)

const (
	eName = "event"
	dName = "data"
)

type Params struct {
	RetryTimeout time.Duration
}

type Event struct {
	Event string
	Data  string
}

func (e *Event) addToMessage(str string) {
	e.Data += str
}

var DefaultParams = Params{
	RetryTimeout: 5 * time.Second,
}

type Requester func() (*http.Request, error)

type Client struct {
	Requester    Requester
	EventChannel chan Event
	ErrorChannel chan error
}

func (c *Client) Run(p *Params) {
	if p == nil {
		p = &DefaultParams
	}
	go c.run(p)
}

func (c *Client) RunSync(p *Params) {
	if p == nil {
		p = &DefaultParams
	}
	go c.run(p)
}

func (c *Client) run(p *Params) {
	client := &http.Client{}

	reqG := c.Requester
	ev := c.EventChannel
	errCh := c.ErrorChannel

	for {
		req, err := reqG()
		req.Header.Set("Accept", "text/event-stream")
		if err != nil {
			errCh <- err
			time.Sleep(p.RetryTimeout)
			continue
		}
		res, err := client.Do(req)
		if err != nil {
			errCh <- err
		} else {
			parse(res, ev, errCh)
		}
		time.Sleep(p.RetryTimeout)
	}
}

func cleanBytes(byt []byte) string {
	return string(bytes.TrimSpace(byt))
}

func parse(res *http.Response, evCh chan Event, errCh chan error) {
	br := bufio.NewReader(res.Body)
	defer res.Body.Close()

	currEvent := &Event{}

	// We want to send out the events if there is 250 milliseconds
	// without any other data coming through. However, the ReadBytes
	// function blocks. Therefore, we need to use goroutines
	// and channels.
	byteCh := make(chan []byte)

	go func() {
		for {
			bs, err := br.ReadBytes('\n')
			if err != nil {
				errCh <- err
				return
			}
			if len(bs) < 2 {
				continue
			}
			byteCh <- bs
		}
	}()

	for {
		select {
		case <-time.After(250 * time.Millisecond):
			if currEvent.Event != "" {
				send(evCh, *currEvent)
				currEvent = &Event{}
			}
		case bs := <-byteCh:
			spl := bytes.SplitN(bs, []byte{':'}, 2)
			if len(spl) < 2 {
				if currEvent.Event != "" {
					currEvent.addToMessage(string(bs))
				}
				continue
			}
			switch cleanBytes(spl[0]) {
			case eName:
				if currEvent.Event != "" {
					send(evCh, *currEvent)
				}
				currEvent = &Event{
					Event: cleanBytes(spl[1]),
				}
			case dName:
				currEvent.addToMessage(string(spl[1]))
			}
		}
	}
}

func send(evCh chan Event, ev Event) {
	if ev.Event == "" {
		return
	}
	ev.Data = strings.TrimSpace(ev.Data)
	evCh <- ev
}
