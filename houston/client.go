package houston

import (
	"io/ioutil"
	"net/http"

	"fmt"

	"encoding/json"

	"github.com/astronomerio/event-router/config"
	"github.com/astronomerio/event-router/pkg"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Client the HTTPClient used to communicate with the HoustonAPI
type Client struct {
	HTTPClient *pkg.HTTPClient
	APIUrl     string
}

type HoustonResponse struct {
	Raw  *http.Response
	Body []byte
}

// NewHoustonClient returns a new Client with the logger and HTTP client setup.
func NewHoustonClient(HTTPClient *pkg.HTTPClient) *Client {
	return &Client{
		HTTPClient: HTTPClient,
	}
}

type GraphQLQuery struct {
	Query string `json:"query"`
}

type QueryOrganizationResponse struct {
	Data struct {
		Organization struct {
			Sources []struct {
				Clickstream []struct {
					ID      string `json:"id"`
					Code    string `json:"code"`
					Enabled bool   `json:"enabled"`
				} `json:"clickstream"`
			} `json:"sources"`
		} `json:"organization"`
	} `json:"data"`
}

type VerifyTokenResponse struct {
	Data struct {
		VerifyToken struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
		} `json:"verifyToken"`
	} `json:"data"`
}

type CreateTokenResponse struct {
	Data struct {
		CreateToken struct {
			Success bool   `json:"success"`
			Token   string `json:"token"`
			Message string `json:"message"`
		} `json:"createToken"`
	} `json:"data"`
}

var (
	log = logrus.WithField("package", "houston")

	queryOrganization = `
	query organizations {
	  organization(orgId:"%s"){
		sources {
		  clickstream {
			id
			code
			enabled
		  }
		}
	  }
	}`

	createToken = `
	mutation createToken {
	  createToken(username:"%s", password:"%s") {
		success
		message
		token
	  }
	}`

	verifyToken = `
	mutation VerifyUserToken {
	  verifyToken(token: "%s"){
	    success
	    message
	  }
	}`

	authToken string
)

func (c *Client) GetOrganizationIntegrations(appId string) ([]string, error) {
	logger := log.WithField("function", "GetOrganizationIntegrations")
	logger.WithField("appId", appId).Debug("Entered GetOrganizationIntegrations")

	query := fmt.Sprintf(queryOrganization, appId)
	authKey, err := c.getAuthorizationKey()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting organization integrations")
	}
	houstonResponse, err := c.queryHouston(query, authKey)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting organization integrations")
	}
	var queryResponse QueryOrganizationResponse
	if err = json.Unmarshal(houstonResponse.Body, queryResponse); err != nil {
		return nil, errors.Wrap(err, "Error decoding get organization integrations response")
	}
	integrationsMap := make(map[string]bool)
	integrations := make([]string, 0)
	for _, src := range queryResponse.Data.Organization.Sources {
		for _, clickStream := range src.Clickstream {
			if clickStream.Enabled && !integrationsMap[clickStream.Code] {
				integrationsMap[clickStream.Code] = clickStream.Enabled
				integrations = append(integrations, clickStream.Code)
			}
		}
	}
	return integrations, nil
}

func (c *Client) queryHouston(query string, authKey string) (HoustonResponse, error) {
	logger := log.WithField("function", "QueryHouston")
	logger.WithField("query", query).Debug("Querying Houston")
	doOpts := &pkg.DoOptions{
		Data: GraphQLQuery{query},
		Headers: map[string]string{
			"Accept": "application/json",
		},
	}
	if len(authKey) != 0 {
		doOpts.Headers["authorization"] = authKey
	}
	var response HoustonResponse
	httpResponse, err := c.HTTPClient.Do("POST", c.APIUrl, doOpts)
	if err != nil {
		return response, err
	}
	defer httpResponse.Body.Close()
	// strings.NewReader(jsonStream)
	body, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return response, err
	}
	response = HoustonResponse{httpResponse, body}
	return response, nil
}

func (c *Client) getAuthorizationKey() (string, error) {
	// Check for a houston api key
	houstonAPIKey := config.GetString(config.HoustonAPIKeyEnvLabel)
	var authKey string
	if len(houstonAPIKey) != 0 {
		authKey = houstonAPIKey
	} else {
		at, err := c.getToken()
		if err != nil {
			return "", errors.Wrap(err, "Error getting authorization key")
		}
		authToken = at
		authKey = at
	}
	return authKey, nil
}

func (c *Client) getToken() (string, error) {
	// If we have a token, verify it
	if len(authToken) != 0 {
		success, err := c.verifyToken()
		if err != nil {
			return "", errors.Wrap(err, "Error getting token")
		}
		if success {
			return authToken, nil
		}
	}
	query := fmt.Sprintf(createToken,
		config.GetString(config.HoustonUserNameEnvLabel),
		config.GetString(config.HoustonPasswordEnvLabel))
	houstonResponse, err := c.queryHouston(query, "")
	if err != nil {
		return "", errors.Wrap(err, "Error creating token")
	}
	var createTokenRespones CreateTokenResponse
	if err = json.Unmarshal(houstonResponse.Body, createTokenRespones); err != nil {
		return "", errors.Wrap(err, "Error unmarshaling create token response")
	}
	if createTokenRespones.Data.CreateToken.Success {
		return createTokenRespones.Data.CreateToken.Token, nil
	}
	return "", errors.New(createTokenRespones.Data.CreateToken.Message)
}

func (c *Client) verifyToken() (bool, error) {
	query := fmt.Sprintf(verifyToken, authToken)
	queryResponse, err := c.queryHouston(query, "")
	if err != nil {
		return false, errors.Wrap(err, "Error verifying token")
	}
	var verifyResponse VerifyTokenResponse
	if err = json.Unmarshal(queryResponse.Body, verifyResponse); err != nil {
		return false, errors.Wrap(err, "Error unmarshaling verify response")
	}
	return verifyResponse.Data.VerifyToken.Success, nil
}
