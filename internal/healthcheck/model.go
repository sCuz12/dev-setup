package healthcheck

type HealthCheck struct {
	Type string `yaml:"type"` // "http", "tcp", "command"
	URL  string `yaml:"url"`  // for "http"
	Port int    `yaml:"port"` // for "tcp"
	Cmd  string `yaml:"cmd"`  // for "command"
}