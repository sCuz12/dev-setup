package hooks

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/sCuz12/dev-setup/internal/config"
)

func RunHooks (service config.ServiceConfig,trigger string,containerName string) error {
	for _,hook := range service.Hooks{
		
		//skip
		if(trigger != hook.Trigger) {
			continue
		}

		if err := executeHook(service, hook,containerName); err != nil {
            return fmt.Errorf("hook %s failed: %w", hook.Name, err)
        }
	}
	return nil
}

func executeHook(service config.ServiceConfig, hook config.HookConfig, containerName string) error {
    fmt.Printf("Running hook '%s' for service %s inside container %s the command %s \n", hook.Name, service.Name, containerName,hook.Run)
    cmd := exec.Command("docker", "exec", containerName, "sh", "-c", hook.Run)
	cmd.Dir = service.Path
    
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    if err := cmd.Run(); err != nil {
        return fmt.Errorf("hook '%s' failed in container %s: %w", hook.Name, containerName, err)
    }

    fmt.Printf("Hook '%s' completed successfully inside %s.\n", hook.Name, containerName)
    return nil
}