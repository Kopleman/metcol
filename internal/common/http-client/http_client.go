package httpclient

import (
	"errors"
	"io"
	"net/http"
)

type IHTTPClient interface {
	Post(url, contentType string, body io.Reader) ([]byte, error)
}

func (c *HTTPClient) Post(url, contentType string, body io.Reader) ([]byte, error) {
	finalURL := c.BaseURL + url
	var respBody []byte
	res, err := c.client.Post(finalURL, contentType, body)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = errors.Join(err, res.Body.Close())
	}()

	respBody, err = io.ReadAll(res.Body)
	return respBody, err
}

type HTTPClient struct {
	BaseURL string
	client  *http.Client
}

func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		BaseURL: baseURL,
		client:  &http.Client{},
	}
}
