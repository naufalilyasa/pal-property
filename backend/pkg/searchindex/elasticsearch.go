package searchindex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	address    string
	username   string
	password   string
	httpClient *http.Client
}

func NewClient(address, username, password string, httpClient *http.Client) (*Client, error) {
	if strings.TrimSpace(address) == "" {
		return nil, fmt.Errorf("searchindex: elastic address is required")
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{address: strings.TrimRight(address, "/"), username: username, password: password, httpClient: httpClient}, nil
}

func (c *Client) UpsertDocument(ctx context.Context, index, id string, document any) error {
	body, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("searchindex: marshal document: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s/%s/_doc/%s", c.address, index, id), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("searchindex: create upsert request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.username != "" || c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("searchindex: upsert request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("searchindex: upsert failed with status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) EnsureIndex(ctx context.Context, index string, mapping any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, fmt.Sprintf("%s/%s", c.address, index), nil)
	if err != nil {
		return fmt.Errorf("searchindex: create head request: %w", err)
	}
	if c.username != "" || c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("searchindex: head request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	if resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("searchindex: head failed with status %d", resp.StatusCode)
	}
	return c.CreateIndex(ctx, index, mapping)
}

func (c *Client) CreateIndex(ctx context.Context, index string, mapping any) error {
	body, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("searchindex: marshal mapping: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s/%s", c.address, index), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("searchindex: create index request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.username != "" || c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("searchindex: create index request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("searchindex: create index failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}
	return nil
}

func (c *Client) DeleteIndex(ctx context.Context, index string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", c.address, index), nil)
	if err != nil {
		return fmt.Errorf("searchindex: create delete index request: %w", err)
	}
	if c.username != "" || c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("searchindex: delete index request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("searchindex: delete index failed with status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) RecreateIndex(ctx context.Context, index string, mapping any) error {
	if err := c.DeleteIndex(ctx, index); err != nil {
		return err
	}
	return c.CreateIndex(ctx, index, mapping)
}

func (c *Client) DeleteDocument(ctx context.Context, index, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("%s/%s/_doc/%s", c.address, index, id), nil)
	if err != nil {
		return fmt.Errorf("searchindex: create delete request: %w", err)
	}
	if c.username != "" || c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("searchindex: delete request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("searchindex: delete failed with status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) Search(ctx context.Context, index string, query any, target any) error {
	body, err := json.Marshal(query)
	if err != nil {
		return fmt.Errorf("searchindex: marshal search query: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/%s/_search", c.address, index), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("searchindex: create search request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.username != "" || c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("searchindex: search request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("searchindex: search failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("searchindex: decode search response: %w", err)
	}
	return nil
}
