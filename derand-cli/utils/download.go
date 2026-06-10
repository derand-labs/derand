package utils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var client = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	},
}

func DownloadOrOpenFile(ctx context.Context, url string) (*os.File, error) {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return os.Open(url)
	}

	fname, err := DownloadFile(ctx, url)
	if err != nil {
		return nil, err
	}

	return os.Open(fname)
}

func DownloadFile(ctx context.Context, url string) (string, error) {
	out, err := os.CreateTemp("", "*.download")
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer out.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %s", resp.Status)
	}

	buf := make([]byte, 4*1024*1024)

	if _, err := io.CopyBuffer(out, resp.Body, buf); err != nil {
		return "", fmt.Errorf("download file: %w", err)
	}

	if err := out.Sync(); err != nil {
		return "", fmt.Errorf("sync file: %w", err)
	}

	if err := out.Close(); err != nil {
		return "", fmt.Errorf("close file: %w", err)
	}

	return out.Name(), nil
}
