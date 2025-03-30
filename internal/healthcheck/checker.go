package healthcheck

import (
	"fmt"
	"time"
)

type HealthChecker interface {
	Check() error
}

func WaitForService(checker HealthChecker) error {
	timeout := time.After(60 * time.Second)
    ticker := time.NewTicker(3 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-timeout:
            return fmt.Errorf("timeout waiting for service")
        case <-ticker.C:
            if err := checker.Check(); err == nil {
                fmt.Println("Service is healthy")
                return nil
            } else {
                fmt.Println("Still waiting...")
            }
        }
    }
}