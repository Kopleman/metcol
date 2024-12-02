package httpclient

import (
	"io"
	"net/http"

	"github.com/Kopleman/metcol/internal/agent/config"
	"github.com/Kopleman/metcol/internal/common/log"
)

func (c *HTTPClient) Post(url, contentType string, body io.Reader) ([]byte, error) {
	finalURL := c.BaseURL + url
	var respBody []byte
	res, err := c.client.Post(finalURL, contentType, body)
	if err != nil {
		return nil, err
	}

	defer func() {
		if bodyParseErr := res.Body.Close(); bodyParseErr != nil {
			c.logger.Error(bodyParseErr)
		}
	}()

	respBody, err = io.ReadAll(res.Body)
	return respBody, err
}

type HTTPClient struct {
	BaseURL string
	client  *http.Client
	logger  log.Logger
}

func NewHTTPClient(cfg *config.Config, logger log.Logger) *HTTPClient {
	baseURL := `http://` + cfg.EndPoint.String()

	return &HTTPClient{
		BaseURL: baseURL,
		client:  &http.Client{},
		logger:  logger,
	}
}
