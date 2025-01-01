package httpclient

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"

	"github.com/Kopleman/metcol/internal/agent/config"
	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/log"
)

func (c *HTTPClient) Post(url, contentType string, body io.Reader) ([]byte, error) {
	finalURL := c.BaseURL + url
	var respBody []byte

	req, err := http.NewRequest(http.MethodPost, finalURL, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set(common.ContentType, contentType)
	req.Header.Set(common.AcceptEncoding, "gzip")

	res, respErr := c.client.Do(req)
	if respErr != nil {
		return nil, fmt.Errorf("failed to send post req to '%s': %w", finalURL, respErr)
	}

	if res.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("failed to send post req to '%s': status code %d", finalURL, res.StatusCode)
	}

	gz, gzipErr := gzip.NewReader(res.Body)
	if gzipErr != nil {
		return nil, fmt.Errorf("failed to decompress response: %w", err)
	}
	defer func() {
		if gzErr := gz.Close(); gzErr != nil {
			c.logger.Error(gzErr)
		}
	}()

	defer func() {
		if bodyParseErr := res.Body.Close(); bodyParseErr != nil {
			c.logger.Error(bodyParseErr)
		}
	}()

	respBody, err = io.ReadAll(gz)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w", err)
	}

	return respBody, nil
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
