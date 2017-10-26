package houston

import (
	"io/ioutil"
	"net/http"

	"fmt"

	"encoding/json"

	"github.com/astronomerio/clickstream-event-router/config"
	"github.com/astronomerio/clickstream-event-router/pkg"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type HoustonClient interface {
	GetIntegrations(appId string) (*map[string]string, error)
	GetAuthorizationKey() (string, error)
}

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
func NewHoustonClient(HTTPClient *pkg.HTTPClient, apiURL string) *Client {
	return &Client{
		HTTPClient: HTTPClient,
		APIUrl:     apiURL,
	}
}

type GraphQLQuery struct {
	Query string `json:"query"`
}

type VerifyTokenResponse struct {
	Data struct {
		VerifyToken struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
		} `json:"verifyToken"`
	} `json:"data"`
	Errors []Errors `json:"errors"`
}

type CreateTokenResponse struct {
	Data struct {
		CreateToken struct {
			Success bool   `json:"success"`
			Token   string `json:"token"`
			Message string `json:"message"`
		} `json:"createToken"`
	} `json:"data"`
	Errors []Errors `json:"errors"`
}

type QuerySourcesResponse struct {
	Data struct {
		Sources []struct {
			Clickstream []struct {
				Name    string `json:"name"`
				Topic   string `json:"topic"`
				Enabled bool   `json:"enabled"`
			} `json:"clickstream"`
		} `json:"sources"`
	} `json:"data"`
	Errors []Errors `json:"errors"`
}

type Errors struct {
	Message   string `json:"message"`
	Locations []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations"`
	Path []string `json:"path"`
}

var (
	log = logrus.WithField("package", "houston")

	querySources = `
	query sources {
	  sources(id:"%s") {
		clickstream{
			name
		  topic: code
		  enabled
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

// GetIntegrations will get the enabled integrations from Houston
func (c *Client) GetIntegrations(appId string) (*map[string]string, error) {
	logger := log.WithField("function", "GetIntegrations")
	logger.WithField("appId", appId).Debug("Entered GetIntegrations")

	query := fmt.Sprintf(querySources, appId)
	authKey, err := c.GetAuthorizationKey()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting auth key")
	}
	houstonResponse, err := c.queryHouston(query, authKey)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting source integrations")
	}
	var queryResponse QuerySourcesResponse
	if err = json.Unmarshal(houstonResponse.Body, &queryResponse); err != nil {
		return nil, errors.Wrap(err, "Error decoding query sources response")
	}
	integrationsMap := make(map[string]string)
	for _, src := range queryResponse.Data.Sources {
		for _, clickStream := range src.Clickstream {
			if clickStream.Enabled && len(integrationsMap[clickStream.Name]) == 0 {
				integrationsMap[clickStream.Name] = clickStream.Topic
			}
		}
	}
	return &integrationsMap, nil
}

func (c *Client) queryHouston(query string, authKey string) (HoustonResponse, error) {
	logger := log.WithField("function", "QueryHouston")
	doOpts := &pkg.DoOptions{
		Data: GraphQLQuery{query},
		Headers: map[string]string{
			"Accept": "application/json",
		},
	}
	logger.WithField("query", doOpts.Data).Debug("Querying Houston")
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

func (c *Client) GetAuthorizationKey() (string, error) {
	// Check for a houston api key
	houstonAPIKey := config.GetString(config.HoustonAPIKey)
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
			// Dont' want to return, try to get a new key
			log.Infof("Error validating auth key: %v", err)
		}
		if success {
			return authToken, nil
		}
	}
	query := fmt.Sprintf(createToken,
		config.GetString(config.HoustonUserName),
		config.GetString(config.HoustonPassword))
	houstonResponse, err := c.queryHouston(query, "")
	if err != nil {
		return "", errors.Wrap(err, "Error creating token")
	}
	var createTokenRespones CreateTokenResponse
	if err = json.Unmarshal(houstonResponse.Body, &createTokenRespones); err != nil {
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
	if err = json.Unmarshal(queryResponse.Body, &verifyResponse); err != nil {
		return false, errors.Wrap(err, "Error unmarshaling verify response")
	}
	return verifyResponse.Data.VerifyToken.Success, nil
}
