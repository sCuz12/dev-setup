package healthcheck

import (
	"fmt"
	"net/http"
	"time"
)

type HttpHealthCheck struct {
	URL string
} 

func (h *HttpHealthCheck) Check ()error {
	client := &http.Client{
        Timeout: 2 * time.Second,
    }
    resp, err := client.Get(h.URL)
    if err != nil {
        return fmt.Errorf("HTTP request failed: %w", err)
    }
    defer resp.Body.Close()

    fmt.Println(resp)
    if resp.StatusCode >= 200 && resp.StatusCode < 300 {
        return nil // Healthy
    }
    return fmt.Errorf("unhealthy status code: %d", resp.StatusCode)
}