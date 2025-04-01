package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/sCuz12/dev-setup/internal/config"
	"github.com/sCuz12/dev-setup/internal/healthcheck"
	"github.com/sCuz12/dev-setup/internal/hooks"
	"github.com/spf13/cobra"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"gopkg.in/yaml.v3"
)

type Config struct {
	GlobalEnv map[string]string `yaml:"globalEnv"`
	Services  []config.ServiceConfig   `yaml:"services"`
	Docker    []DockerConfig    `yaml:"docker"`
	Hooks     []config.HookConfig      `yaml:"hooks"`
}

type DockerConfig struct {
	Name        string                   `yaml:"name"`
	Image       string                   `yaml:"image"`
	Build       Build                    `yaml:"build"`
	Ports       []string                 `yaml:"volumes"`
	Volumes     []string                 `yaml:"ports"`
	Env         map[string]string        `yaml:"env"`
	HealthCheck *healthcheck.HealthCheck `yaml:"healthCheck"`
}
type Build struct {
	Context    string
	Dockerfile string
}
var rootCmd = &cobra.Command{
	Use:   "dev-setup",
	Short: "A tool for spinning up local microservices",
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the local environment",
	RunE: func(cmd *cobra.Command, args []string) error {
		// only, _ := cmd.Flags().GetStringSlice("only")
		// skipDocker, _ := cmd.Flags().GetBool("skip-docker")
		// clean, _ := cmd.Flags().GetBool("clean")
		// return runInit(only, skipDocker, clean)
		return runInit()
	},
}

func main() {
	rootCmd.AddCommand(initCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runInit() error {
	//load config
	cfg, err := LoadConfig()

	if err != nil {
		fmt.Println(err)
		return err
	}

	resolvedServices, err := resolveServices(cfg)

	if err != nil {
		return err
	}
	// startDockerContainer(cfg.Docker[1])
	//clone repos
	if err := cloneRepos(resolvedServices); err != nil {
		return err
	}

	// setup their docker-compose
	runMicroserviceDockerCompose(resolvedServices)

	// Health Check
	// err = performHealthCheck(resolvedServices)
	
	// Step 6: Run post-up hooks
	  for _, svc := range resolvedServices {
		containerName, err := resolveContainerNameFromService(svc.Name)

		if err != nil {
			return fmt.Errorf(err.Error())
		}

        if err := hooks.RunHooks(svc, "post-up",containerName); err != nil {
            return fmt.Errorf("post-up hooks failed: %w", err)
        }
    }

	if err != nil {
		return err
	}

	// Start Microservices
	return nil
}

func cloneRepos(resolvedServices []config.ServiceConfig) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(resolvedServices))

	for _, svc := range resolvedServices {
		svc := svc 
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := cloneSingleRepo(svc); err != nil {
				errChan <- err
			}
		}()
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		return err // fail fast
	}
	return nil
}

func cloneSingleRepo(service config.ServiceConfig) error {
	fmt.Printf("Cloning the repository. %s\n", service.Repo)
	// Path to your private key, e.g. ~/.ssh/id_rsa
	// If you have a passphrase, pass it as the last parameter here.
	auth, err := ssh.NewPublicKeysFromFile("git", os.ExpandEnv("$HOME/.ssh/id_rsa"), "")
	if err != nil {
		return fmt.Errorf("could not load public keys: %w", err)
	}

	_, err = git.PlainClone(service.Path, false, &git.CloneOptions{
		URL:      service.Repo,
		Progress: os.Stdout,
		Auth:     auth,
	})

	if err != nil {
		return fmt.Errorf("could not clone repo %s: %w", service.Repo, err)
	}
	return nil
}

func LoadConfig() (*Config, error) {
	yfile, err := os.ReadFile("dev-setup.yaml")

	if err != nil {
		fmt.Println(err)
	}

	var config Config
	err = yaml.Unmarshal(yfile, &config)
	fmt.Println(config)
	return &config, nil
}

func resolveServices(cfg *Config) ([]config.ServiceConfig, error) {
	serviceMap := make(map[string]config.ServiceConfig)

	for _, svc := range cfg.Services {
		serviceMap[svc.Name] = svc
	}

	return cfg.Services, nil
}

func startDockerContainer(d DockerConfig) error {
	color.Magenta("Starting container with name %s", d.Name)
	var imageToRun string
	// 1) Build if there's a build context
	if d.Build.Context != "" {
		buildArgs := []string{"build", "-t", d.Name + ":latest"}
		if d.Build.Dockerfile != "" {
			buildArgs = append(buildArgs, "-f", d.Build.Dockerfile)
		}
		buildArgs = append(buildArgs, d.Build.Context)

		cmdBuild := exec.Command("docker", buildArgs...)
		cmdBuild.Stdout = os.Stdout
		cmdBuild.Stderr = os.Stderr

		if err := cmdBuild.Run(); err != nil {
			return fmt.Errorf("failed to build image for %s: %w", d.Name, err)
		}

		// If we built an image, we use "<name>:latest" for `docker run`
		imageToRun = d.Name + ":latest"
	} else {
		// 2) Pull if no build context
		cmdPull := exec.Command("docker", "pull", d.Image)
		cmdPull.Stdout = os.Stdout
		cmdPull.Stderr = os.Stderr

		if err := cmdPull.Run(); err != nil {
			return fmt.Errorf("failed to pull image %s: %w", d.Image, err)
		}
		// If we're not building, we use the user-provided image name
		imageToRun = d.Image
	}

	// 3) Compose `docker run` arguments
	runArgs := []string{"run", "-d", "--name", d.Name}

	for _, p := range d.Ports {
		runArgs = append(runArgs, "-p", p)
	}
	for k, v := range d.Env {
		runArgs = append(runArgs, "-e", fmt.Sprintf("%s=%s", k, v))
	}
	//Handle Volumes
	for _, vol := range d.Volumes {
		runArgs = append(runArgs, "-v", vol)
	}

	// **Always** run either the pulled or just-built image
	runArgs = append(runArgs, imageToRun)
	cmdRun := exec.Command("docker", runArgs...)
	cmdRun.Stdout = os.Stdout
	cmdRun.Stderr = os.Stderr

	if err := cmdRun.Run(); err != nil {
		return fmt.Errorf("failed to start container %s: %w", d.Name, err)
	}
	return nil
}

func runMicroserviceDockerCompose(services []config.ServiceConfig) error {
	for _, srv := range services {
		if srv.ComposeFile != "" {

			//Auto discovery of compose file
			err := runDockerCompose(srv)
			if err != nil {
				fmt.Errorf("failed to run compose for service %s: %w", srv.Name, err)
			}
		}
	}
	return nil
}

func runDockerCompose(singleService config.ServiceConfig) error {
	//get the filename
	composePaths, err := findComposeFiles(singleService.Path, singleService.ComposeFile)

	if err != nil {
		return fmt.Errorf("Err: %w", err)
	}

	if len(composePaths) == 0 {
		return fmt.Errorf("no compose file found for service %s", singleService.Name)
	}

	extractedComposePath := composePaths[0]

	filenameOnly := filepath.Base(extractedComposePath)
	//execute the container docker-compose up
	cmd := exec.Command("docker", "compose", "-p", singleService.Name, "-f", filenameOnly, "up", "-d")
	cmd.Dir = singleService.Path

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Starting Docker Compose for %s using %s\n", singleService.Name, singleService.Path)

	fmt.Println("Running docker-compose from directory:", cmd.Dir)
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func findComposeFiles(rootPath string, composerFileName string) ([]string, error) {

	var composeFiles []string

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && (strings.HasPrefix(d.Name(), composerFileName) && strings.HasSuffix(composerFileName, ".yml")) {
			composeFiles = append(composeFiles, path)
		}
		return nil
	})

	return composeFiles, err
}
func performHealthCheck(services []config.ServiceConfig) error {
	for _, svc := range services {
		fmt.Printf("Performing Health check for service: %s", svc.Name)

		if svc.HealthCheck != nil {
			containerName, err := resolveContainerNameFromService(svc.Name)
			doctor, err := healthcheck.NewHealthChecker(svc.HealthCheck, containerName)

			err = doctor.Check()
			fmt.Println(err)

			if err != nil {
				return fmt.Errorf("invalid health check for service %s: %w", svc.Name, err)
			}

			fmt.Printf("Waiting for %s to become healthy...\n", svc.Name)
			// if err := doctor.WaitForService(); err != nil {
			// 	return fmt.Errorf("service %s failed health check: %w", svc.Name, err)
			// }
		}
	}
	return nil
}

func resolveContainerNameFromService(serviceName string) (string, error) {
	// Use docker ps to get containers filtered by name
	cmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("name=%s", serviceName), "--format", "{{.Names}}")

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute docker ps: %w", err)
	}

	containerName := strings.TrimSpace(string(output))
	if containerName == "" {
		return "", fmt.Errorf("no container found matching service name: %s", serviceName)
	}

	return containerName, nil
}
