package marathon

import (
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/astronomerio/asds/postgres"
	"github.com/astronomerio/asds/provisioner"
)

func TestProvision(t *testing.T) {
	marathonClient := &Client{
		dbClient:       postgres.NewTestClient(),
		log:            logrus.New().WithField("test", "test"),
		marathonConfig: &Config{},
		marathonClient: Marathon{},
	}

	req := &provisioner.ProvisionRequest{
		OrgID: "12345",
	}
	resp, err := marathonClient.Provision(req)
	if err != nil {
		t.Error(err)
	}
	if resp == nil {
		t.Error("Empty Response")
	}
	if resp.OrgID != "12345" {
		t.Error("OrgID doesn't match")
	}
}
