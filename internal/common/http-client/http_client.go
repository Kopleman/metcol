package httpclient

import (
	"fmt"
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
		return nil, fmt.Errorf("failed to send post req to '%s': %w", finalURL, err)
	}

	defer func() {
		if bodyParseErr := res.Body.Close(); bodyParseErr != nil {
			c.logger.Error(bodyParseErr)
		}
	}()

	respBody, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w", err)
	}

	return respBody, err
}

type HTTPClient struct {
	logger  log.Logger
	client  *http.Client
	BaseURL string
}

func NewHTTPClient(cfg *config.Config, logger log.Logger) *HTTPClient {
	baseURL := `http://` + cfg.EndPoint.String()

	return &HTTPClient{
		BaseURL: baseURL,
		client:  &http.Client{},
		logger:  logger,
	}
}
