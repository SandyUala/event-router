package sse

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sync"
)

const (
	eventMarker = "event"
	dataMarker  = "data"
	idMarker    = "id"

	emptyLine = "\n"
)

const (
	StatusDisconnected = iota
	StatusConnected
	StatusReconnecting
	StatusRetryExhausted
	StatusFatalError
)

type Status struct {
	Code  int
	Error error
}

type SSEClient struct {
	uri         string
	httpClient  *http.Client
	httpHeaders map[string]string

	eventLock     sync.RWMutex
	eventChannels map[string]EventChan

	delivery   EventChan
	statusChan StatusChan

	shutdown chan struct{}
}

type Event struct {
	ID    string
	Event string
	Data  []byte
}

type EventChan chan *Event
type StatusChan chan Status

func NewClient() *SSEClient {
	return &SSEClient{
		httpClient:    &http.Client{},
		eventChannels: map[string]EventChan{},
		delivery:      make(EventChan),
		shutdown:      make(chan struct{}),
		statusChan:    make(StatusChan),
	}
}

var DefaultClient = NewClient()

func Subscribe(event string) EventChan {
	return DefaultClient.Subscribe(event)
}

func WithHTTPClient(client *http.Client) *SSEClient {
	return DefaultClient.WithHTTPClient(client)
}

func WithHTTPHeaders(headers map[string]string) *SSEClient {
	return DefaultClient.WithHTTPHeaders(headers)
}

func WithURI(uri string) *SSEClient {
	return DefaultClient.WithURI(uri)
}

func Events() StatusChan {
	return DefaultClient.Events()
}

func Listen() error {
	return DefaultClient.Listen()
}

func Close() {
	DefaultClient.Close()
}

func (c *SSEClient) WithHTTPClient(client *http.Client) *SSEClient {
	c.httpClient = client
	return c
}

func (c *SSEClient) WithHTTPHeaders(headers map[string]string) *SSEClient {
	c.httpHeaders = headers
	return c
}

func (c *SSEClient) WithURI(uri string) *SSEClient {
	c.uri = uri
	return c
}

func (c *SSEClient) Subscribe(event string) EventChan {
	return c.addListener(event)
}

func (c *SSEClient) addListener(event string) EventChan {
	c.eventLock.Lock()
	defer c.eventLock.Unlock()

	c.eventChannels[event] = make(EventChan)
	return c.eventChannels[event]
}

func (c *SSEClient) maybeDeliver(e *Event) {
	c.eventLock.RLock()
	defer c.eventLock.RUnlock()

	if c, ok := c.eventChannels[e.Event]; ok {
		c <- e
	}
}

func (c *SSEClient) Events() StatusChan {
	return c.statusChan
}

func (c *SSEClient) Listen() error {
	req, err := http.NewRequest(http.MethodGet, c.uri, nil)
	if err != nil {
		return err
	}

	for k, v := range c.httpHeaders {
		req.Header.Add(k, v)
	}

	req.Header.Add("Accept", "text/event-stream")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("status code not 200: %d", res.StatusCode)
	}

	br := bufio.NewReader(res.Body)
	defer res.Body.Close()

	delim := []byte{':', ' '}

	go func() {
		e := &Event{}
		for {
			bs, err := br.ReadBytes('\n')
			if err != nil && err != io.EOF {
				c.statusChan <- Status{
					Code:  StatusFatalError,
					Error: err,
				}
				return
			}

			spl := bytes.Split(bs, delim)

			if len(spl) == 1 && string(spl[0]) == emptyLine {
				c.delivery <- e
				e = &Event{}
				continue
			}

			switch string(spl[0]) {
			case idMarker:
				e.ID = string(bytes.TrimSpace(spl[1]))
			case eventMarker:
				e.Event = string(bytes.TrimSpace(spl[1]))
			case dataMarker:
				e.Data = bytes.TrimSpace(spl[1])
			}

			if err == io.EOF {
				break
			}
		}
	}()

outer:
	for {
		select {
		case e := <-c.delivery:
			c.maybeDeliver(e)
		case <-c.shutdown:
			res.Body.Close()
			break outer
		}
	}

	return nil
}

func (c *SSEClient) Close() {
	c.shutdown <- struct{}{}
}
