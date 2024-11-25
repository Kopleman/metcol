package http_client

import (
	"errors"
	"io"
	"net/http"
)

type IHttpClient interface {
	Post(url, contentType string, body io.Reader) ([]byte, error)
}

func (c *HttpClient) Post(url, contentType string, body io.Reader) ([]byte, error) {
	finalUrl := c.BaseURL + url
	var respBody []byte
	res, err := c.client.Post(finalUrl, contentType, body)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = errors.Join(err, res.Body.Close())
	}()

	respBody, err = io.ReadAll(res.Body)
	return respBody, err
}

type HttpClient struct {
	BaseURL string
	client  *http.Client
}

func NewHttpClient(baseURL string) IHttpClient {
	return &HttpClient{
		BaseURL: baseURL,
		client:  &http.Client{},
	}
}
