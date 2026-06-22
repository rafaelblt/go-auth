package testutil

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type HTTPClient struct {
	baseURL string
}

func NewHTTPClient(baseURL string) (*HTTPClient, error) {
	if baseURL == "" {
		return nil, errors.New("base url empty")
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base url: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("base url %q must include scheme and host (e.g. http://localhost:8080)", baseURL)
	}
	baseURL = strings.TrimSuffix(baseURL, "/")
	client := HTTPClient{baseURL: baseURL}
	return &client, nil
}

func (c *HTTPClient) do(t *testing.T, method, path string, body any) *http.Response {
	t.Helper()

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	var reader io.Reader
	switch b := body.(type) {
	case nil:
		// no body
	case io.Reader:
		reader = b
		if rc, ok := b.(io.ReadCloser); ok {
			defer rc.Close()
		}
	case string:
		reader = strings.NewReader(b)
	case []byte:
		reader = bytes.NewReader(b)
	default:
		data, err := json.Marshal(body)
		require.NoError(t, err, "marshal body failed")
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reader)
	require.NoError(t, err, "request creation failed")

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err, "client request do failed")

	return resp
}

func (c *HTTPClient) Get(t *testing.T, path string) *http.Response {
	t.Helper()
	return c.do(t, http.MethodGet, path, nil)
}

func (c *HTTPClient) Post(t *testing.T, path string, body any) *http.Response {
	t.Helper()
	return c.do(t, http.MethodPost, path, body)
}
