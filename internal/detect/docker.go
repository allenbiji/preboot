package detect

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/allenbiji/preboot/internal/model"
	"gopkg.in/yaml.v3"
)

// the struct is used to hold the services block and the ports that they use
type composeConfig struct {
	Services map[string]struct {
		Ports []string `yaml:"ports"`
	} `yaml:"services"`
}

// extractHostPort handles various Docker Compose port string formats
func extractHostPort(mapping string) string {
	parts := strings.Split(mapping, ":")

	if len(parts) == 2 {
		// Format: "8080:80"
		return parts[0]
	} else if len(parts) == 3 {
		// Format: "127.0.0.1:8080:80"
		return parts[1]
	}

	// format that is not recognized yet
	return ""
}

func detectDockerCompose() []model.CheckConfig{
	var checks []model.CheckConfig

	var targetFile string
	if _, err := os.Stat("docker-compose.yml"); err == nil {
		targetFile = "docker-compose.yml"
	} else if _, err := os.Stat("compose.yaml"); err == nil {
		targetFile = "compose.yaml"
	} else {
		// Neither exists, exit early and return an empty slice
		return checks
	}

	checks = append(checks, model.CheckConfig{
		Name: "docker-installed",
		Type: model.TypeCommandExists,
		Severity: model.SeverityBlocker,
		Options: map[string]string{"command": "docker"},
		Message: "Docker is required for running local infrastructure",
	})

	data, err := os.ReadFile(targetFile)
	if err != nil {
		return checks
	}

	var parsed composeConfig
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		fmt.Fprintf(os.Stderr, "warn: could not parse %s: %v\n", targetFile, err)
		return checks
	}

	serviceNames := make([]string, 0, len(parsed.Services))
	for name := range parsed.Services {
		serviceNames = append(serviceNames, name)
	}
	slices.Sort(serviceNames)

	for _, serviceName := range serviceNames {
		service := parsed.Services[serviceName]
		for _, portMapping := range service.Ports {
			hostPort := extractHostPort(portMapping)

			if hostPort == "" {
				continue
			}
			if _, err := strconv.Atoi(hostPort); err != nil {
				continue // non-numeric (env var reference, named port) — skip
			}

			checks = append(checks, model.CheckConfig{
				Name:     "port-free-" + hostPort,
				Type:     model.TypePortFree,
				Severity: model.SeverityBlocker,
				Options:  map[string]string{"port": hostPort},
				Message:  "Port " + hostPort + " is in use. Stop background services so " + serviceName + " can start.",
			})
		}
	}

	return checks
}


