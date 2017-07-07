package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	Base string
}

func (c *Client) MakeRequest(url string, output interface{}, headers map[string]string) (*http.Response, error) {
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

	if res.StatusCode > 400 {
		return nil, fmt.Errorf("Invalid status code observed (%d) for URL: %s", res.StatusCode, url)
	}

	if output != nil {
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(res.Body)

		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(buf.Bytes(), &output)
	}

	return res, err
}
