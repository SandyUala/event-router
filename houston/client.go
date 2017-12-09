package houston

import (
	"io/ioutil"
	"net/http"

	"fmt"

	"encoding/json"

	"time"

	"sync"

	"github.com/astronomerio/event-router/config"
	"github.com/astronomerio/event-router/pkg"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type HoustonClient interface {
	GetIntegrations(appId string) (*map[string]string, error)
	GetAuthorizationToken() (string, error)
}

// Client the HTTPClient used to communicate with the HoustonAPI
type Client struct {
	HTTPClient *pkg.HTTPClient
	APIUrl     string
	authToken  *houstonAuthToken
	mutex      *sync.Mutex
}

type houstonResponse struct {
	Raw  *http.Response
	Body []byte
}

type houstonAuthToken struct {
	token   string
	expDate time.Time
}

// NewHoustonClient returns a new Client with the logger and HTTP client setup.
func NewHoustonClient(HTTPClient *pkg.HTTPClient, apiURL string) *Client {
	c := &Client{
		HTTPClient: HTTPClient,
		APIUrl:     apiURL,
		mutex:      &sync.Mutex{},
	}
	c.authToken = &houstonAuthToken{}
	if len(config.GetString(config.HoustonAPIKey)) != 0 {
		c.authToken.token = config.GetString(config.HoustonAPIKey)
	}
	return c
}

type GraphQLQuery struct {
	Query string `json:"query"`
}

type VerifyTokenResponse struct {
	Data struct {
		VerifyToken struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
			Decoded struct {
				Exp time.Time `json:"exp"`
			} `json:"decoded"`
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
			Decoded struct {
				Exp time.Time `json:"exp"`
			} `json:"decoded"`
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
	    decoded{
	      exp
	    }
	  }
	}`

	verifyToken = `
	mutation VerifyUserToken {
	  verifyToken(token: "%s"){
	    success
	    message
	    decoded{
	      exp
	    }
	  }
	}`
)

// GetIntegrations will get the enabled integrations from Houston
func (c *Client) GetIntegrations(appId string) (*map[string]string, error) {
	logger := log.WithField("function", "GetIntegrations")
	logger.WithField("appId", appId).Debug("Entered GetIntegrations")

	query := fmt.Sprintf(querySources, appId)
	token, err := c.GetAuthorizationToken()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting auth key")
	}
	houstonResponse, err := c.queryHouston(query, token)
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
	logger.WithField("integrations", integrationsMap).Debug("Returning Integrations")
	return &integrationsMap, nil
}

func (c *Client) GetAuthorizationToken() (string, error) {
	logger := log.WithField("function", "GetAuthorizationToken")
	logger.Debug("Entered GetAuthorizationToken")
	c.mutex.Lock()
	defer c.mutex.Unlock()
	// If we have an auth token, and exp date is good, return that
	if len(c.authToken.token) != 0 {
		if c.authToken.expDate.Before(time.Now()) {
			return c.authToken.token, nil
		}
		// Auth token has an invalid time, lets see if we can verify it!
		logger.Debug("Token expired, validating")
		verifyToken, err := c.verifyToken()
		if err != nil {
			return "", errors.Wrap(err, "Error getting authorization token")
		}
		if verifyToken.Data.VerifyToken.Success {
			// Update the exp date
			c.authToken.expDate = verifyToken.Data.VerifyToken.Decoded.Exp
			return c.authToken.token, nil
		}
	}
	logger.Debug("Token invalid, creating")
	// Not a valid token!  Create a new one!
	createResponse, err := c.createToken()
	if err != nil {
		return "", errors.Wrap(err, "Error getting authorization token")
	}
	// Update the auth token
	c.authToken.token = createResponse.Data.CreateToken.Token
	c.authToken.expDate = createResponse.Data.CreateToken.Decoded.Exp
	return c.authToken.token, nil
}

func (c *Client) createToken() (*CreateTokenResponse, error) {
	if len(config.GetString(config.HoustonUserName)) == 0 ||
		len(config.GetString(config.HoustonPassword)) == 0 {
		return nil, errors.New("Houston username/password not provided, can't create new token")
	}
	query := fmt.Sprintf(createToken,
		config.GetString(config.HoustonUserName),
		config.GetString(config.HoustonPassword))

	houstonResponse, err := c.queryHouston(query, "")
	if err != nil {
		return nil, errors.Wrap(err, "Error creating token")
	}
	var createTokenRespones CreateTokenResponse
	if err = json.Unmarshal(houstonResponse.Body, &createTokenRespones); err != nil {
		return nil, errors.Wrap(err, "Error unmarshaling create token response")
	}
	return &createTokenRespones, nil
}

func (c *Client) verifyToken() (*VerifyTokenResponse, error) {
	query := fmt.Sprintf(verifyToken, c.authToken.token)
	queryResponse, err := c.queryHouston(query, "")
	if err != nil {
		return nil, errors.Wrap(err, "Error verifying token")
	}
	var verifyResponse VerifyTokenResponse
	if err = json.Unmarshal(queryResponse.Body, &verifyResponse); err != nil {
		return nil, errors.Wrap(err, "Error unmarshaling verify response")
	}
	return &verifyResponse, nil
}

func (c *Client) queryHouston(query string, authKey string) (houstonResponse, error) {
	logger := log.WithField("function", "QueryHouston")
	doOpts := &pkg.DoOptions{
		Data: GraphQLQuery{query},
		Headers: map[string]string{
			"Accept": "application/json",
		},
	}
	logger.Debug("Querying Houston")
	if len(authKey) != 0 {
		doOpts.Headers["authorization"] = authKey
	}
	var response houstonResponse
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
	response = houstonResponse{httpResponse, body}
	return response, nil
}
