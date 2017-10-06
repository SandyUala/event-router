package houston

import (
	"io/ioutil"
	"net/http"

	"github.com/astronomerio/event-router/pkg"
	"github.com/sirupsen/logrus"
)

// Client the HTTPClient used to communicate with the HoustonAPI
type Client struct {
	HTTPClient *pkg.HTTPClient
	APIUrl     string
}

type HoustonResponse struct {
	Raw  *http.Response
	Body string
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

var (
	log = logrus.WithField("package", "houston")

	querySelfRequest = `
	query self {
	  self {
	    id
	    username
	  }
	}`
)

func (c *Client) QueryHouston(query string) (HoustonResponse, error) {
	logger := log.WithField("function", "QueryHouston")
	logger.WithField("query", query).Debug("Querying Houston")
	doOpts := &pkg.DoOptions{
		Data: GraphQLQuery{query},
		Headers: map[string]string{
			"Accept": "application/json",
		},
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

	response = HoustonResponse{httpResponse, string(body)}

	return response, nil
}
