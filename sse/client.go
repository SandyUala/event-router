package sse

import (
	"time"

	"github.com/astronomerio/clickstream-event-router/houston"
	"github.com/astronomerio/sse"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("package", "sse")

type SSE interface {
	Subscribe(channel string, handler func(event *sse.Event))
}

type Client struct {
	houstonClient  houston.HoustonClient
	sseUri         string
	shouldShutdown bool
	shutdownChan   chan struct{}
	restartChan    chan struct{}
}

func NewSSEClient(sseUri string, houstonClient houston.HoustonClient, shutdownChannel chan struct{}) (*Client, error) {
	return &Client{
		houstonClient: houstonClient,
		shutdownChan:  shutdownChannel,
		sseUri:        sseUri,
		restartChan:   make(chan struct{}),
	}, nil
}

func (c *Client) Subscribe(channel string, handler func(event *sse.Event)) {
	logger := log.WithField("function", "start")
	go func() {
		// For loop so we restart
		for !c.shouldShutdown {
			logger.Debug("Getting SSE Auth Token")
			authToken, err := c.houstonClient.GetAuthorizationToken()
			if err != nil {
				logger.Error(err)
				// Sleep for 5 seconds
				time.Sleep(time.Second * 5)
				continue
			}
			sse.WithURI(c.sseUri)
			sse.WithHTTPHeaders(map[string]string{
				"authorization": authToken,
			})
			exitChan := make(chan struct{})
			// Start listening
			c.listen(channel, handler, exitChan)
			logger.Info("Starting SSE Client")
			if err := sse.Listen(); err != nil {
				logger.Error(err)
				close(exitChan)
				continue
			}
			// Sleep for 5 seconds
			time.Sleep(time.Second * 5)
		}
	}()
}

func (c *Client) listen(channel string, handler func(event *sse.Event), exitChannel chan struct{}) {
	logger := log.WithField("function", "listen")
	logger.Info("Starting SSE Listeners")
	go func() {
		clickstreamEvents := sse.Subscribe(channel)
		//heartbeat := sse.Subscribe("heartbeat")
		events := sse.Events()
		for {
			select {
			case event := <-clickstreamEvents:
				if event == nil {
					logger.Warn("Received nil event")
					sse.Close()
					time.Sleep(time.Second * 5)
					return
				}
				if event.Event == channel {
					handler(event)
				}
			//case e := <-heartbeat:
			//	if e != nil {
			//		logger.Debug(e.Event)
			//	}
			case e := <-events:
				logger.Error(e.Error)
				// Close the SSE channel
				sse.Close()
				return
			case <-c.shutdownChan:
				c.shouldShutdown = true
				sse.Close()
				return
			case <-exitChannel:
				c.shouldShutdown = true
				sse.Close()
				return
			}
		}
	}()
}
