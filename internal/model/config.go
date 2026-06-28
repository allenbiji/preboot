package model

// these are the levels of severity the system can handle
type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityBlocker Severity = "blocker"
)

// these are the enums for the checks that the system can handle
type CheckType string

const (
	TypeEnvExists       CheckType = "env_exists"
	TypeCommandExists   CheckType = "command_exists"
	TypeFileExists      CheckType = "file_exists"
	TypeDirectoryExists CheckType = "directory_exists"
	TypeHttpReachable   CheckType = "http_reachable"
	TypeTcpReachable    CheckType = "tcp_reachable"
	TypePortFree        CheckType = "port_free"
)

// this is the type for each of the checks
type CheckConfig struct {
	Name     string            `yaml:"name"`
	Type     CheckType         `yaml:"type"`
	Severity Severity          `yaml:"severity"`
	Options  map[string]string `yaml:"options"`
	Message  string            `yaml:"message"`
	Why      string            `yaml:"why"`
	Fix      string            `yaml:"fix"`
}

// this is the type for the global preboot config in its yaml file
type PrebootConfig struct {
	Version  int                    `yaml:"version"`
	Defaults map[string]interface{} `yaml:"defaults"` 
	Checks   []CheckConfig          `yaml:"checks"`
}
