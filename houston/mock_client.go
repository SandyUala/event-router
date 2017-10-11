package houston

type MockClient struct {
	Integrations map[string]string
}

func (c *MockClient) GetOrganizationIntegrations(appId string) (map[string]string, error) {
	return c.Integrations, nil
}
