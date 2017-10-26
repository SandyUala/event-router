package houston

type MockClient struct {
	Integrations *map[string]string
}

func (c *MockClient) GetIntegrations(appId string) (*map[string]string, error) {
	return c.Integrations, nil
}

func (c *MockClient) GetAuthorizationKey() (string, error) {
	return "", nil
}
