
package coordinator

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type NodeClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewNodeClient(baseURL string) (*NodeClient, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid node URL: %w", err)
	}

	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("node URL must include scheme and host")
	}

	return &NodeClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (c *NodeClient) PutChunk(
	ctx context.Context,
	id string,
	src io.Reader,
) error {
	endpoint := c.baseURL + "/chunks/" + url.PathEscape(id)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPut,
		endpoint,
		src,
	)
	if err != nil {
		return fmt.Errorf("create PUT request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send PUT request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return responseError(resp)
	}

	return nil
}

func (c *NodeClient) GetChunk(
	ctx context.Context,
	id string,
) (io.ReadCloser, error) {
	endpoint := c.baseURL + "/chunks/" + url.PathEscape(id)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create GET request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send GET request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		err := responseError(resp)
		resp.Body.Close()
		return nil, err
	}

	return resp.Body, nil
}

func (c *NodeClient) DeleteChunk(
	ctx context.Context,
	id string,
) error {
	endpoint := c.baseURL + "/chunks/" + url.PathEscape(id)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		endpoint,
		nil,
	)
	if err != nil {
		return fmt.Errorf("create DELETE request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send DELETE request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return responseError(resp)
	}

	return nil
}

func responseError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	return fmt.Errorf(
		"storage node returned status %d: %s",
		resp.StatusCode,
		strings.TrimSpace(string(body)),
	)
}
