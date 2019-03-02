package api

import "github.com/cj123/canijailbreak.com/model"

const CanIJailbreakURL = "https://canijailbreak.com/"

// NewIPSWClient creates an IPSWClient. If client == nil, http.DefaultClient is used.
func NewCanIJailbreakClient(apiBase string, httpClient HTTPClient) *CanIJailbreakClient {
	return &CanIJailbreakClient{
		client: newHTTPWrapper(apiBase, httpClient),
	}
}

type CanIJailbreakClient struct {
	client *ipswHTTPWrapper
}

func (c *CanIJailbreakClient) GetJailbreaks() (*model.Jailbreaks, error) {
	var jbs *model.Jailbreaks

	resp, _, err := c.client.makeRequest("jailbreaks.json", nil)

	if err != nil {
		return nil, err
	}

	err = parseJSON(resp, &jbs)

	return jbs, err
}
