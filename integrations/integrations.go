package integrations

import (
	"sync"

	"encoding/json"

	"time"

	"github.com/astronomerio/clickstream-event-router/config"
	"github.com/astronomerio/clickstream-event-router/houston"
	"github.com/astronomerio/clickstream-event-router/pkg/prom"
	"github.com/astronomerio/sse"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var (
	log             = logrus.WithField("package", "integrations")
	integrationsMap = NewMap()
	syncMap         = make(map[string]*sync.RWMutex)
	globalLock      sync.Mutex
	cacheTimer      time.Timer
)

type Integrations interface {
	GetIntegrations(appId string) (*map[string]string, error)
	UpdateIntegrationsForApp(appId string) error
	EventListener(event *sse.Event)
}

/*
 Map holding a cache of enabled integrations
*/

type Map struct {
	integrations map[string]*map[string]string
}

func NewMap() *Map {
	return &Map{integrations: make(map[string]*map[string]string)}
}

func (m *Map) Get(key string) *map[string]string {
	globalLock.Lock()
	defer globalLock.Unlock()
	lock, ok := syncMap[key]
	if !ok {
		lock = &sync.RWMutex{}
		syncMap[key] = lock
	}
	lock.RLock()
	defer lock.RUnlock()
	return m.integrations[key]

}

func (m *Map) Put(key string, value *map[string]string) {
	m.integrations[key] = value
}

type Client struct {
	houstonClient houston.HoustonClient
	shutdownChan  chan struct{}
}

func NewClient(houstonClient houston.HoustonClient, shutdownChan chan struct{}) *Client {
	client := &Client{
		houstonClient: houstonClient,
		shutdownChan:  shutdownChan,
	}
	if !config.GetBool(config.DisableCacheTTL) {
		client.StartTTL()
	}
	return client
}

func (c *Client) StartTTL() {
	ttl := config.GetInt(config.CacheTTLMin)
	log.Infof("Setting Integration Cache TTL to %dm", ttl)
	timer := time.NewTimer(time.Minute * time.Duration(ttl))
	go func() {
		for {
			select {
			case <-c.shutdownChan:
				return
			case <-timer.C:
				c.resetCache()
			}
		}
	}()
}

func (c *Client) resetCache() {
	log.Debug("Resetting Integration Cache")
	// Get a global lock so we can kill the cache
	globalLock.Lock()
	defer globalLock.Unlock()
	integrationsMap = NewMap()
	// Empty the sync map as well, remove stale app ids
	syncMap = make(map[string]*sync.RWMutex)
}

func (c *Client) GetIntegrations(appId string) (*map[string]string, error) {
	integrations := integrationsMap.Get(appId)
	if integrations == nil {
		prom.IntegrationCacheMiss.With(prometheus.Labels{"appId": appId}).Inc()
		log.Debugf("Populating integrations for appId %s", appId)
		if err := c.UpdateIntegrationsForApp(appId); err != nil {
			return nil, errors.Wrap(err, "Error getting integration")
		}
	}
	//log.WithFields(logrus.Fields{"integrations": integrations, "appId": appId}).Debug("Returning integrations")
	return integrations, nil
}

func (c *Client) getIntegrationsFromHouston(appId string) (*map[string]string, error) {
	integrations, err := c.houstonClient.GetIntegrations(appId)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return integrations, nil
}

func (c *Client) UpdateIntegrationsForApp(appId string) error {
	prom.IntegrationCacheUpdate.With(prometheus.Labels{"appId": appId}).Inc()
	ints, err := c.houstonClient.GetIntegrations(appId)
	if err != nil {
		return errors.Wrap(err, "Error updating integrations")
	}
	// Call get on the map to ensure a lock for the API was created
	integrationsMap.Get(appId)
	syncMap[appId].Lock()
	integrationsMap.Put(appId, ints)
	syncMap[appId].Unlock()
	return nil
}

type SSEMessage struct {
	AppID string `json:"appId"`
}

func (c *Client) EventListener(event *sse.Event) {
	logger := log.WithField("function", "EventListener")
	message := &SSEMessage{}
	if err := json.Unmarshal(event.Data, message); err != nil {
		logger.Error(err)
		return
	}
	prom.SSEClickstreamMessagesReceived.Inc()
	c.UpdateIntegrationsForApp(message.AppID)
	logger.Infof("AppID %s integrations updated", message.AppID)
}

// Mock Client for testing
type MockClient struct {
}

func (c *MockClient) GetIntegrations(appId string) (*map[string]string, error) {
	return nil, nil
}
func (c *MockClient) UpdateIntegrationsForApp(appId string) error {
	return nil
}
func (c *MockClient) EventListener(event *sse.Event) {

}
