package healthcheck

import (
	"fmt"
	"os/exec"
)

type CommandHealthCheck struct {
	Container string
	Command string
}

func (c *CommandHealthCheck) Check() error {
	fmt.Printf("Container : %s",c.Container)
	cmd := exec.Command("docker", "exec", c.Container, "sh", "-c", c.Command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %s - %w", string(output), err)
	}
	return nil
}