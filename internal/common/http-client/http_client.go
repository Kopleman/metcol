package httpclient

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"

	"github.com/Kopleman/metcol/internal/agent/config"
	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/davecgh/go-spew/spew"
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

	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send post req to '%s': %w", finalURL, err)
	}

	gz, err := gzip.NewReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress response: %w", err)
	}
	defer gz.Close() //nolint:all // defer

	defer func() {
		if bodyParseErr := res.Body.Close(); bodyParseErr != nil {
			c.logger.Error(bodyParseErr)
		}
	}()
	spew.Dump(res.Header)

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
