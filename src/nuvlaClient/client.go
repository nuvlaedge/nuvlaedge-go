package nuvlaClient

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/cookiejar"
	"nuvlaedge-go/src/common/resources"
)

type NuvlaClient struct {
	EndPoint string
	Insecure bool
	WhoAmI   string // WhoAmI: what CiMiResource I am logging in as. Default should be NuvlaBox
	Client   *http.Client
}

func NewNuvlaClient(endpoint string, insecure bool) *NuvlaClient {
	jar, err := cookiejar.New(nil)
	if err != nil {
		fmt.Println("Error creating cookie jar:", err)
		return nil
	}
	return &NuvlaClient{
		EndPoint: endpoint,
		Insecure: insecure,
		WhoAmI:   "nuvlabox",
		Client:   &http.Client{Jar: jar},
	}
}

func (c *NuvlaClient) LogIn(cred *resources.Credentials) bool {

	loginCredentials := cred.FormatCredentialsPayload()

	loginPayload := map[string]resources.CredentialPayload{"template": loginCredentials}
	resp, _ := c.Post(loginPayload, "/api/session")

	if resp.StatusCode != 201 {
		log.Warnf("NuvlaEdge Log in failed to create a New session with  error code %d", resp.StatusCode)
		return false
	} else {
		log.Debug("NuvlaEdge LogIn success")
		return true
	}
}

// Post executes the post http method
func (c *NuvlaClient) Post(data interface{}, endpointPath string) (*http.Response, error) {
	endpoint := c.EndPoint + endpointPath

	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(payload))
	req.Header = map[string][]string{
		"Content-Type": {"application/json"},
		"Accept":       {"application/json"},
	}

	response, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	log.Debugf("Post %s response status code %d", endpointPath, response.StatusCode)
	return response, nil
}

func (c *NuvlaClient) Put(data interface{}, endPointPath string) (*http.Response, error) {
	endPoint := c.EndPoint + endPointPath

	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("PUT", endPoint, bytes.NewBuffer(payload))
	req.Header = map[string][]string{
		"Content-Type": {"application/json"},
		"Accept":       {"application/json"},
	}

	response, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	log.Debugf("Put %s response status code %d", endPointPath, response.StatusCode)
	return response, nil
}

func (c *NuvlaClient) Get(endPointPath string) (*http.Response, error) {
	endPoint := c.EndPoint + endPointPath

	resp, err := c.Client.Get(endPoint)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *NuvlaClient) Authenticated() bool {
	return false
}
