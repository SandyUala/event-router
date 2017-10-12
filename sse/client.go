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

func NewSSEClient(sseUrl string) *Client {
	return &Client{
		client: sse.NewClient(sseUrl),
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
	logger.Info("Subscribed to SS3 Stream %s", stream)
	eventChan := make(chan *sse.Event)
	err := c.client.SubscribeChan(stream, eventChan)
	if err != nil {
		logger.Error(err)
		return
	}

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	run := true
	go func() {
		for run {
			select {
			case sig := <-sigchan:
				logger.Infof("Consumer caught signal %v: terminating", sig)
				run = false
			case ev := <-eventChan:
				logger.WithFields(logrus.Fields{
					"event": string(ev.Event),
					"data":  string(ev.Data),
				}).Debug("Received Event")
				handler(ev.Event, ev.Data)
			}
		}
	}()
	logger.Infof("Closing SSE Stream %s", stream)
}
