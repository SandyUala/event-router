package integrations

import (
	"sync"

	"github.com/astronomerio/event-router/houston"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	log             = logrus.WithField("package", "integrations")
	integrationsMap = NewMap()
)

/*
 Map holding a cache of enabled integrations
*/

type Map struct {
	integrations map[string][]string
	sync.RWMutex
}

func NewMap() *Map {
	return &Map{integrations: make(map[string][]string)}
}

func (m *Map) Get(key string) []string {
	m.RLock()
	defer m.RUnlock()
	return m.integrations[key]
}

func (m *Map) Put(key string, value []string) {
	m.Lock()
	defer m.Unlock()
	m.integrations[key] = value
}

type Client struct {
	houstonClient *houston.Client
}

func NewClient(houstonClient *houston.Client) *Client {
	return &Client{houstonClient: houstonClient}
}

func (c *Client) GetIntegrations(appId string) []string {
	integrations := integrationsMap.Get(appId)
	if integrations == nil {
		if integrations = c.getIntegrationsFromHouston(appId); integrations != nil {
			integrationsMap.Put(appId, integrations)
		}
	}
	return integrations
}

func (c *Client) getIntegrationsFromHouston(appId string) []string {
	ints, err := c.houstonClient.GetOrganizationIntegrations(appId)
	if err != nil {
		log.Error(err)
		return nil
	}
	return ints
}

func (c *Client) UpdateIntegrationsForApp(appId string) error {
	ints, err := c.houstonClient.GetOrganizationIntegrations(appId)
	if err != nil {
		return errors.Wrap(err, "Error updating integrations")
	}
	integrationsMap.Put(appId, ints)
	return nil
}