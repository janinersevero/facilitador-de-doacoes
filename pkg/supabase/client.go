package supabase

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	url        string
	key        string
	bucket     string
	httpClient *http.Client
}

func NewClient(url, key, bucket string) *Client {
	return &Client{
		url:        url,
		key:        key,
		bucket:     bucket,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) UploadFile(ctx context.Context, fileName string, data []byte, contentType string) (string, error) {
	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", c.url, c.bucket, fileName)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.key)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "true")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("supabase upload failed with status %d", resp.StatusCode)
	}

	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", c.url, c.bucket, fileName)
	return publicURL, nil
}
