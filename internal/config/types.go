package config

type HookConfig struct {
	Name    string `yaml:"name"`
	Trigger string `yaml:"trigger"` 
	Run string `yaml:"run"`
}