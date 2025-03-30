package healthcheck

import "fmt"


func NewHealthChecker(hc *HealthCheck,containerName string) (HealthChecker, error) {
    switch hc.Type {
    case "http":
        return &HttpHealthCheck{URL: hc.URL}, nil
    case "command":
        return &CommandHealthCheck{
            Container: containerName, 
            Command:   hc.Cmd,
        }, nil
	//add more if we want to extend more types
    default:
        return nil, fmt.Errorf("unsupported health check type: %s", hc.Type)
    }
}