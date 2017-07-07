package ipsw

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/cj123/ranger"
)

var DefaultClient HTTPClient = &http.Client{
	Timeout: 30 * time.Second,
}

// HTTPClient is a wrapper for client
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
	Get(string) (*http.Response, error)
	Head(string) (*http.Response, error)
	Post(string, string, io.Reader) (*http.Response, error)
	PostForm(string, url.Values) (*http.Response, error)
}

func bufferedDownload(file *zip.File, writer io.Writer) error {
	rc, err := file.Open()

	if err != nil {
		return err
	}

	defer rc.Close()

	_, err = io.Copy(writer, rc)

	return err
}

func DownloadFile(resource, file string, w io.Writer) error {
	u, err := url.Parse(resource)

	if err != nil {
		return err
	}

	reader, err := ranger.NewReader(
		&ranger.HTTPRanger{
			URL:    u,
			Client: DefaultClient,
		},
	)

	if err != nil {
		return err
	}

	zipReader, err := zip.NewReader(reader, reader.Length())

	if err != nil {
		return err
	}

	for _, f := range zipReader.File {
		if f.Name == file {
			return bufferedDownload(f, w)
		}
	}

	return fmt.Errorf("pwn: file '%s' not found in resource '%s'", file, resource)
}
