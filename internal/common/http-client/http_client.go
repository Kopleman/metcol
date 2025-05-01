// Package httpclient is wrapper around net/http client.
package httpclient

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/Kopleman/metcol/internal/agent/config"
	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/log"
)

// getOutboundIP fetches current IP address.
func (c *HTTPClient) getOutboundIP() (net.IP, error) {
	if c.outboundIP != nil {
		return c.outboundIP, nil
	}
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to outbound IP: %w", err)
	}
	defer conn.Close() //nolint:errcheck //safe to ignore

	localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return nil, errors.New("failed to get local IP address")
	}
	c.outboundIP = localAddr.IP
	return c.outboundIP, nil
}

// Post perform post request to dest url.
func (c *HTTPClient) Post(url, contentType string, bodyBytes []byte) ([]byte, error) {
	body := bytes.NewBuffer(bodyBytes)
	finalURL := c.BaseURL + url
	var respBody []byte

	req, err := http.NewRequest(http.MethodPost, finalURL, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set(common.ContentType, contentType)
	req.Header.Set(common.AcceptEncoding, "gzip")

	ip, oErr := c.getOutboundIP()
	if oErr == nil {
		req.Header.Set("X-Real-IP", ip.String())
	}

	bodyHash := c.calcHashForBody(bodyBytes)
	if bodyHash != "" {
		req.Header.Set(common.HashSHA256, bodyHash)
	}

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

func (c *HTTPClient) calcHashForBody(bodyBytes []byte) string {
	if len(c.key) == 0 {
		return ""
	}
	if len(bodyBytes) == 0 {
		return ""
	}

	h := hmac.New(sha256.New, c.key)
	h.Write(bodyBytes)
	hash := h.Sum(nil)
	hashString := base64.StdEncoding.EncodeToString(hash)

	return hashString
}

type HTTPClient struct {
	logger     log.Logger
	client     *http.Client
	outboundIP net.IP
	BaseURL    string
	key        []byte
}

const defaultRetryCount = 3

func NewHTTPClient(cfg *config.Config, logger log.Logger) *HTTPClient {
	baseURL := `http://` + cfg.EndPoint.String()

	transport := NewRetryableTransport(logger, defaultRetryCount)

	return &HTTPClient{
		BaseURL: baseURL,
		client: &http.Client{
			Transport: transport,
		},
		logger: logger,
		key:    []byte(cfg.Key),
	}
}
