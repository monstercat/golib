package sse

import (
	"bufio"
	"bytes"
	"io"
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

func Run(reqG Requester, ev chan Event, err chan error, p *Params) {
	if p == nil {
		p = &DefaultParams
	}
	go run(reqG, ev, err, p)
}

func RunSync(reqG Requester, ev chan Event, err chan error, p *Params) {
	if p == nil {
		p = &DefaultParams
	}
	run(reqG, ev, err, p)
}

func run(reqG Requester, ev chan Event, errCh chan error, p *Params) {
	client := &http.Client{}

	for {
		req, err := reqG()
		if err != nil {
			errCh <- err
			time.Sleep(p.RetryTimeout)
			continue
		}
		res, err := client.Do(req)
		if err != nil {
			errCh <- err
		}else {
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

	for {
		bs, err := br.ReadBytes('\n')
		if err != nil && err != io.EOF {
			errCh <- err
			return
		}
		if err == io.EOF {
			time.Sleep(100 * time.Millisecond)
		}
		if len(bs) < 2 {
			continue
		}
		spl := bytes.Split(bs, []byte{':'})
		if len(spl) < 2 {
			if currEvent.Event != "" && len(bs) > 0 {
				currEvent.addToMessage(string(bs))
			}

			continue
		}

		switch cleanBytes(spl[0]) {
		case eName:
			if currEvent.Event != "" {
				currEvent.Data = strings.TrimSpace(currEvent.Data)
				evCh <- *currEvent
				currEvent = &Event{}
			}
			currEvent.Event = cleanBytes(spl[1])
		case dName:
			currEvent.addToMessage(string(spl[1]))
		}
	}

}
