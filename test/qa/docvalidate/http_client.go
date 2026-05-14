package main

import (
	"net/http"
	"time"
)

type httpClient struct {
	timeout time.Duration
}

func (c *httpClient) head(raw string) (int, error) {
	client := &http.Client{Timeout: c.timeout}
	resp, err := client.Head(raw)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}
