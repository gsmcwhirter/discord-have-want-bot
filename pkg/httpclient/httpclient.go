package httpclient

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/gsmcwhirter/eso-discord/pkg/logging"
)

// HTTPClient TODOC
type HTTPClient interface {
	Get(context.Context, string, *http.Header) (*http.Response, error)
	Post(context.Context, string, *http.Header, io.Reader) (*http.Response, error)
}

type dependencies interface {
	Logger() log.Logger
}

type httpClient struct {
	deps    dependencies
	client  *http.Client
	headers http.Header
}

// NewHTTPClient TODOC
func NewHTTPClient(deps dependencies) HTTPClient {
	return httpClient{
		deps:    deps,
		client:  &http.Client{},
		headers: http.Header{},
	}
}

func addHeaders(to *http.Header, from http.Header) {
	for k, v := range from {
		to.Del(k)
		if len(v) > 1 {
			for _, v2 := range v {
				to.Add(k, v2)
			}
		} else if len(v) == 1 {
			to.Set(k, v[0])
		}
	}
}

func (c httpClient) Get(ctx context.Context, url string, headers *http.Header) (*http.Response, error) {
	logger := logging.WithContext(ctx, c.deps.Logger())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	addHeaders(&req.Header, c.headers)
	if headers != nil {
		addHeaders(&req.Header, *headers)
	}

	level.Debug(logger).Log(
		"message", "http get start",
		"url", url,
		"headers",
	)
	start := time.Now()
	resp, err := c.client.Do(req)
	level.Debug(logger).Log(
		"message", "http get complete",
		"duration_ns", time.Since(start).Nanoseconds(),
		"status_code", resp.StatusCode,
	)

	return resp, err
}

func (c httpClient) Post(ctx context.Context, url string, headers *http.Header, body io.Reader) (*http.Response, error) {
	logger := logging.WithContext(ctx, c.deps.Logger())
	reqHeaders := http.Header{}
	addHeaders(&reqHeaders, c.headers)
	if headers != nil {
		addHeaders(&reqHeaders, *headers)
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	level.Debug(logger).Log(
		"message", "http post start",
		"url", url,
	)

	start := time.Now()
	resp, err := c.client.Do(req)

	level.Debug(logger).Log(
		"message", "http post complete",
		"duration_ns", time.Since(start).Nanoseconds(),
		"status_code", resp.StatusCode,
	)

	return resp, err
}
