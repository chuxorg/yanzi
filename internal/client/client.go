package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/chuxorg/yanzi/internal/core/model"
)

// Client is a minimal HTTP client for the Yanzi Library API.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// IntentRecord mirrors the server v0 schema.
type IntentRecord = model.IntentRecord

// VerifyResponse is returned by the /verify endpoint.
type VerifyResponse struct {
	ID           string  `json:"id"`
	Valid        bool    `json:"valid"`
	StoredHash   string  `json:"stored_hash"`
	ComputedHash string  `json:"computed_hash"`
	PrevHash     string  `json:"prev_hash"`
	Error        *string `json:"error"`
}

// ChainResponse is returned by the /chain endpoint.
type ChainResponse struct {
	HeadID       string         `json:"head_id"`
	Length       int            `json:"length"`
	Intents      []IntentRecord `json:"intents"`
	MissingLinks []string       `json:"missing_links,omitempty"`
}

// ListResponse is returned by the /intents endpoint.
type ListResponse struct {
	Intents []IntentRecord `json:"intents"`
}

// CreateIntentRequest is the payload for POST /v0/intents.
type CreateIntentRequest struct {
	Author     string          `json:"author"`
	SourceType string          `json:"source_type"`
	Title      string          `json:"title,omitempty"`
	Prompt     string          `json:"prompt"`
	Response   string          `json:"response"`
	Meta       json.RawMessage `json:"meta,omitempty"`
	PrevHash   string          `json:"prev_hash,omitempty"`
}

// New creates a client using the provided base URL.
func New(baseURL string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// CreateIntent posts a new intent record.
func (c *Client) CreateIntent(ctx context.Context, req CreateIntentRequest) (IntentRecord, error) {
	var out IntentRecord
	if err := c.doJSON(ctx, http.MethodPost, "/v0/intents", req, &out); err != nil {
		return out, err
	}
	return out, nil
}

// VerifyIntent calls GET /v0/intents/{id}/verify.
func (c *Client) VerifyIntent(ctx context.Context, id string) (VerifyResponse, error) {
	var out VerifyResponse
	path := fmt.Sprintf("/v0/intents/%s/verify", id)
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &out); err != nil {
		return out, err
	}
	return out, nil
}

// ChainIntent calls GET /v0/intents/{id}/chain.
func (c *Client) ChainIntent(ctx context.Context, id string) (ChainResponse, error) {
	var out ChainResponse
	path := fmt.Sprintf("/v0/intents/%s/chain", id)
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &out); err != nil {
		return out, err
	}
	return out, nil
}

// ListIntents calls GET /v0/intents.
func (c *Client) ListIntents(ctx context.Context, author, source string, limit int, metaFilters map[string]string) (ListResponse, error) {
	var out ListResponse
	params := url.Values{}
	if author != "" {
		params.Set("author", author)
	}
	if source != "" {
		params.Set("source", source)
	}
	for key, value := range metaFilters {
		params.Set("meta_"+key, value)
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	path := "/v0/intents"
	if len(params) > 0 {
		path = path + "?" + params.Encode()
	}
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &out); err != nil {
		return out, err
	}
	return out, nil
}

// GetIntent calls GET /v0/intents/{id}.
func (c *Client) GetIntent(ctx context.Context, id string) (IntentRecord, error) {
	var out IntentRecord
	path := fmt.Sprintf("/v0/intents/%s", id)
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &out); err != nil {
		return out, err
	}
	return out, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, body any, out any) error {
	fullURL, err := c.buildURL(path)
	if err != nil {
		return err
	}

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, reader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(respBody))
		if msg == "" {
			msg = resp.Status
		}
		return fmt.Errorf("server error: %s", msg)
	}

	if out == nil {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func (c *Client) buildURL(path string) (string, error) {
	base, err := url.Parse(c.BaseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base_url: %w", err)
	}
	ref, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}
	return base.ResolveReference(ref).String(), nil
}
