package sse

import (
	"os"
	"os/signal"
	"syscall"

	"time"

	"github.com/astronomerio/clickstream-event-router/config"
	"github.com/astronomerio/clickstream-event-router/houston"
	"github.com/r3labs/sse"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("package", "sse")

type Client struct {
	client         *sse.Client
	houstonClient  houston.HoustonClient
	shouldShutdown bool
	shutdownChan   chan struct{}
}

func NewSSEClient(sseUrl string, houstonCLient houston.HoustonClient) *Client {
	client := sse.NewClient(sseUrl)
	return &Client{
		client:        client,
		houstonClient: houstonCLient,
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
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
		<-sigchan
		c.shutdownChan <- struct{}{}
		c.shouldShutdown = true
	}()

	go func() {
		for {
			// Get the auth token.  We get it before we start listening because if the
			// auth token has become invalidated, we need to get a new one.
			// Check if we are using one from env variable or logging in.
			auth := config.GetString(config.HoustonAPIKeyEnvLabel)
			if len(auth) == 0 {
				// Get the auth token
				a, err := c.houstonClient.GetAuthorizationKey()
				if err != nil {
					logger.WithField("error", err).Error("Error getting auth token")
					os.Exit(1)
				}
				auth = a
			}
			c.client.Headers["authorization"] = auth
			if c.listen(stream, handler) && !c.shouldShutdown {
				logger.Info("SSE Connection lost, reconnecting")
			} else {
				logger.Info("SSE Connection closed, exiting")
				return
			}
			time.Sleep(time.Second * time.Duration(5))
		}
	}()
}

func (c *Client) listen(stream string, handler func(event []byte, data []byte)) bool {
	logger := log.WithField("function", "listen")
	logger.Info("Starting SSE Listener")
	eventChan := make(chan *sse.Event)
	if err := c.client.SubscribeChan(stream, eventChan); err != nil {
		logger.Error(err)
		return true
	}
	logger.Infof("Subscribed to SSE Stream %s", stream)

	event := sse.Event{}
	for {
		select {
		case <-c.shutdownChan:
			logger.Infof("SSE Client terminating...")
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
					continue
				}
				event.Event = ev.Event
			case len(ev.Data) != 0:
				event.Data = ev.Data
				handler(event.Event, event.Data)
				logger.WithFields(logrus.Fields{
					"event": string(event.Event),
					"data":  string(event.Data),
				}).Debug("Received Sent")
				event = sse.Event{}
			}

		}
	}
	return true
}
