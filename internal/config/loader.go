package config

import "github.com/sCuz12/dev-setup/internal/healthcheck"



type ServiceConfig struct {
	Name         string                   `yaml:"name"`
	Repo         string                   `yaml:"repo"`
	Path         string                   `yaml:"path"`
	StartCommand string                   `yaml:"startCommand"`
	DependsOn    []string                 `yaml:"dependsOn"`
	HealthCheck  *healthcheck.HealthCheck `yaml:"healthCheck"`
	Env          map[string]string        `yaml:"env"`
	ComposeFile  string                   `yaml:"compose-file"`
	Hooks		[]HookConfig 		      `yaml:"hooks"`	
}