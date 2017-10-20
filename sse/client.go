package sse

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/r3labs/sse"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("package", "sse")

type Client struct {
	client *sse.Client
}

func NewSSEClient(sseUrl string, authorization string) *Client {
	client := sse.NewClient(sseUrl)
	client.Headers["authorization"] = authorization
	return &Client{
		client: client,
	}
}

// Subscribe will subscribe to the SSE stream and send received events to the handler func
func (c *Client) Subscribe(stream string, handler func(event []byte, data []byte)) {
	logger := log.WithFields(
		logrus.Fields{
			"function": "Subscribe",
			"stream":   stream,
		},
	)

	go func() {
		run := true
		for run {
			run = c.listen(stream, handler)
			if run {
				logger.Info("SSE Connection lost, reconnecting")
			} else {
				logger.Info("SSE Connection closed, exiting")
				return
			}
		}
	}()
}

func (c *Client) listen(stream string, handler func(event []byte, data []byte)) bool {
	logger := log.WithField("function", "listen")
	logger.Info("Starting SSE Listener")
	eventChan := make(chan *sse.Event)
	err := c.client.SubscribeChan(stream, eventChan)
	if err != nil {
		logger.Error(err)
		return true
	}
	logger.Infof("Subscribed to SSE Stream %s", stream)
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	run := true
	event := sse.Event{}
	for run {
		select {
		case sig := <-sigchan:
			logger.Infof("SSE Client caught signal %v: terminating", sig)
			logger.Infof("Closing SSE Stream %s", stream)
			run = false
			return false
		case ev := <-eventChan:
			if ev == nil {
				logger.Debug("Received nil event")
				return true
			}
			switch {
			case len(ev.Event) != 0:
				// If the event is a heart beat, ignore it
				if string(ev.Event) == "heartbeat" {
					logger.Debug("SSE Heartbeat")
					event = sse.Event{}
				} else {
					event.Event = ev.Event
				}
			case len(ev.Data) != 0:
				event.Data = ev.Data
			}
			// If we have data, send it up
			if len(event.Data) != 0 {
				handler(event.Event, event.Data)
				logger.WithFields(logrus.Fields{
					"event": string(event.Event),
					"data":  string(event.Data),
				}).Debug("Received Sent")
			}
			event = sse.Event{}
		}
	}
	return true
}
