package api

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type client struct {
	Base string
}

// MakeRequest makes the http request to a given endpoint, optionally unmarshalling json into the response body.
// note: MakeRequest does not call resp.Body.Close(), this must be done manually
func (c *client) MakeRequest(url string, headers map[string]string) (*http.Response, error) {
	request, err := http.NewRequest("GET", c.Base+url, nil)

	if err != nil {
		return nil, err
	}

	request.Header.Add("Accept", "application/json")

	for key, val := range headers {
		request.Header.Add(key, val)
	}

	res, err := http.DefaultClient.Do(request)

	if err != nil {
		return nil, err
	}

	return res, err
}

func parseJSON(res *http.Response, output interface{}) error {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(res.Body)

	if err != nil {
		return err
	}

	return json.Unmarshal(buf.Bytes(), &output)
}
