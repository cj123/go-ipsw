package api

import (
	"errors"
	"io"
	"strings"

	"honnef.co/go/js/xhr"
)

type HTTPClient interface{}

func newHTTPWrapper(base string, _ HTTPClient) *ipswHTTPWrapper {
	return &ipswHTTPWrapper{
		base: base,
	}
}

type ipswHTTPWrapper struct {
	base string
}

func (h *ipswHTTPWrapper) makeRequest(url string, headers map[string]string) (body io.Reader, statusCode int, err error) {
	r := xhr.NewRequest("GET", h.base+url)
	r.Timeout = 30000 // 30 seconds
	r.ResponseType = xhr.Text

	for key, val := range headers {
		r.SetRequestHeader(key, val)
	}

	err = r.Send(nil)

	if err != nil {
		return nil, 0, err
	}

	if r.Status >= 400 {
		return nil, r.Status, errors.New("http: invalid response observed")
	}

	return strings.NewReader(r.ResponseText), r.Status, nil
}
