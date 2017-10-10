package mesos

import (
	"net/http"

	"net/url"

	"encoding/json"

	"github.com/astronomerio/asds/pkg"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type Client struct {
	log       *logrus.Entry
	mesosURL  string
	mesosRole string
	pi        *pkg.PrometheusInstrumentation
}

func NewClient(config *Config, log *logrus.Logger, pi *pkg.PrometheusInstrumentation) *Client {
	logger := log.WithField("package", "mesos")
	return &Client{
		log:       logger,
		mesosURL:  config.MesosURL,
		mesosRole: config.MesosRole,
		pi:        pi,
	}
}

func (c *Client) GetStateSummary() (*pkg.StateSummary, error) {
	resp, err := c.doHTTP("master/state-summary")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	stateSummary := &pkg.StateSummary{}
	err = json.NewDecoder(resp.Body).Decode(stateSummary)
	if err != nil {
		return nil, err
	}

	return stateSummary, nil
}

func (c *Client) GetAvailableResources() (*pkg.Resources, error) {
	// Get our state summary
	stateSummary, err := c.GetStateSummary()
	if err != nil {
		return nil, err
	}

	availResources := &pkg.Resources{}
	// Loop through the slaves and start gathering stats
	for _, slave := range stateSummary.Slaves {
		if slave.Attributes["role"] == c.mesosRole {
			availResources.CPUs += slave.Resources.CPUs - slave.UsedResources.CPUs
			availResources.Mem += slave.Resources.Mem - slave.UsedResources.Mem
			availResources.Disk += slave.Resources.Disk - slave.UsedResources.Disk
		}
	}

	return availResources, nil
}

func (c *Client) GetMaxResources() (*pkg.Resources, error) {
	// Get our state summary
	stateSummary, err := c.GetStateSummary()
	if err != nil {
		return nil, err
	}

	availResources := &pkg.Resources{}
	// Loop through the slaves and start gathering stats
	for _, slave := range stateSummary.Slaves {
		if slave.Attributes["role"] == c.mesosRole {
			availResources.CPUs += slave.Resources.CPUs
			availResources.Mem += slave.Resources.Mem
			availResources.Disk += slave.Resources.Disk
		}
	}
	return availResources, nil
}

func (c *Client) GetNodeResource() (map[string]*pkg.Resources, error) {
	// Get our state summary
	stateSummary, err := c.GetStateSummary()
	if err != nil {
		return nil, err
	}
	availableResources := make(map[string]*pkg.Resources)
	// Loop through the slaves and start gathering stats
	for _, slave := range stateSummary.Slaves {
		if slave.Attributes["role"] == c.mesosRole {
			rsc := &pkg.Resources{
				CPUs: slave.Resources.CPUs - slave.UsedResources.CPUs,
				Mem:  slave.Resources.Mem - slave.UsedResources.Mem,
				Disk: slave.Resources.Disk - slave.UsedResources.Disk,
			}
			availableResources[slave.Hostname] = rsc
		}
	}
	return availableResources, nil
}

func (c *Client) Metrics() {
	resources, _ := c.GetMaxResources()
	c.pi.ProResourceTotalCPU.Set(resources.CPUs)
	c.pi.ProResourceTotalMem.Set(resources.Mem)
	c.pi.ProResourceTotalDisk.Set(resources.Disk)

	// Node Resources
	nodes, _ := c.GetNodeResource()
	for hostname, node := range nodes {
		c.pi.ProResourceAvailableCPU.With(prometheus.Labels{"node": hostname}).Set(node.CPUs)
		c.pi.ProResourceAvailableMem.With(prometheus.Labels{"node": hostname}).Set(node.Mem)
		c.pi.ProResourceAvailableDisk.With(prometheus.Labels{"node": hostname}).Set(node.Disk)
	}
}

func (c *Client) doHTTP(path string) (*http.Response, error) {
	url, err := url.Parse(c.mesosURL + "/" + path)
	if err != nil {
		return nil, err
	}
	resp, err := http.Get(url.String())
	if err != nil {
		return nil, err
	}

	return resp, nil
}
