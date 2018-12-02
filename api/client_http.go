// +build !js

package api

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
	Get(string) (*http.Response, error)
	Head(string) (*http.Response, error)
	Post(string, string, io.Reader) (*http.Response, error)
	PostForm(string, url.Values) (*http.Response, error)
}

func newHTTPWrapper(base string, client HTTPClient) *ipswHTTPWrapper {
	if client == nil {
		client = http.DefaultClient
	}

	return &ipswHTTPWrapper{
		base:       base,
		httpClient: client,
	}
}

type ipswHTTPWrapper struct {
	base       string
	httpClient HTTPClient
}

func (h *ipswHTTPWrapper) makeRequest(url string, headers map[string]string) (body io.Reader, statusCode int, err error) {
	request, err := http.NewRequest("GET", h.base+url, nil)

	if err != nil {
		return nil, 0, err
	}

	request.Header.Add("Accept", "application/json")

	for key, val := range headers {
		request.Header.Add(key, val)
	}

	res, err := h.httpClient.Do(request)

	if err != nil {
		return nil, 0, err
	}

	defer res.Body.Close()

	buf := new(bytes.Buffer)

	_, err = io.Copy(buf, res.Body)

	if err != nil {
		return nil, 0, err
	}

	return buf, res.StatusCode, err
}
