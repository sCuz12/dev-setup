package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"gopkg.in/yaml.v3"
)

type Config struct {
	GlobalEnv map[string]string `yaml:"globalEnv"`
	Services  []ServiceConfig   `yaml:"services"`
	Docker    []DockerConfig    `yaml:"docker"`
	Hooks     []HookConfig      `yaml:"hooks"`
}

type ServiceConfig struct {
	Name         string            `yaml:"name"`
	Repo         string            `yaml:"repo"`
	Path         string            `yaml:"path"`
	StartCommand string            `yaml:"startCommand"`
	DependsOn    []string          `yaml:"dependsOn"`
	HealthCheck  *HealthCheck      `yaml:"healthCheck"`
	Env          map[string]string `yaml:"env"`
}

type DockerConfig struct {
	Name        string            `yaml:"name"`
	Image       string            `yaml:"image"`
	Build       Build             `yaml:"build"`
	Ports       []string          `yaml:"volumes"`
	Volumes 	[]string		  `yaml:"ports"`
	Env         map[string]string `yaml:"env"`
	HealthCheck *HealthCheck      `yaml:"healthCheck"`
}

type Build struct {
	Context    string
	Dockerfile string
}

type HookConfig struct {
	Name    string `yaml:"name"`
	Trigger string `yaml:"trigger"` // e.g. pre-start, post-start
	Service string `yaml:"service"` // which service/container it relates to
	Command string `yaml:"command"`
}

type HealthCheck struct {
	Type string `yaml:"type"` // "http", "tcp", "command"
	URL  string `yaml:"url"`  // for "http"
	Port int    `yaml:"port"` // for "tcp"
	Cmd  string `yaml:"cmd"`  // for "command"
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
	
	// start docker containers
	// for _,dockerConfig := range cfg.Docker {
	// 	startDockerContainer(dockerConfig)
	// }
	startDockerContainer(cfg.Docker[1])

	//clone repos
	if err := cloneRepos(resolvedServices); err != nil {
		return err
	}

	// Start Microservices
	return nil
}

func cloneRepos(resolvedServices []ServiceConfig) error {
	for _, service := range resolvedServices {
		err := cloneSingleRepo(service)

		if err != nil {
			return err
		}
		fmt.Println("ok")
	}
	return nil
}

func cloneSingleRepo(service ServiceConfig) error {
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

func resolveServices(cfg *Config) ([]ServiceConfig, error) {
	serviceMap := make(map[string]ServiceConfig)

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
