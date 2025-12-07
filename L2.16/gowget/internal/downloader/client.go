package downloader

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
}

func New(timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Response оборачивает ответ сервера
type Response struct {
	Body        io.ReadCloser
	ContentType string
	StatusCode  int
	FinalURL    string // Важно для редиректов
}

func (c *Client) Fetch(url string) (*Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Притворяемся обычным браузером
	req.Header.Set("User-Agent", "GoWget/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error: %w", err)
	}

	if resp.StatusCode >= 400 {
		resp.Body.Close()
		return nil, fmt.Errorf("server returned error status: %d", resp.StatusCode)
	}

	return &Response{
		Body:        resp.Body,
		ContentType: resp.Header.Get("Content-Type"),
		StatusCode:  resp.StatusCode,
		FinalURL:    resp.Request.URL.String(),
	}, nil
}