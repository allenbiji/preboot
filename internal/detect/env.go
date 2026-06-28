package detect

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/allenbiji/preboot/internal/model"
)

func generateEnvChecks(fileName string) []model.CheckConfig {
	var checks []model.CheckConfig

	checks = append(checks, model.CheckConfig{
		Name:     "env-file-exists",
		Type:     model.TypeFileExists,
		Severity: model.SeverityBlocker,
		Options:  map[string]string{"path": ".env"},
		Message:  "You must create a .env file",
		Fix:      "Run cp " + fileName + " .env",
	})

	keys, err := ExtractEnvKeys(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warn: could not read %s: %v\n", fileName, err)
		return checks
	}

	envKeys := make([]string, 0, len(keys))
	for key := range keys {
		envKeys = append(envKeys, key)
	}
	slices.Sort(envKeys)
	for _, key := range envKeys {
		checks = append(checks, model.CheckConfig{
			Name:     strings.ToLower(key) + "-configured",
			Type:     model.TypeEnvExists,
			Severity: model.SeverityBlocker,
			Options:  map[string]string{"key": key},
			Message:  key + " is missing from your environment",
		})
	}

	return checks
}

func detectEnv() []model.CheckConfig {

	if fileExists(".env.example") {
		return generateEnvChecks(".env.example")
	}

	if fileExists(".env.template") {
		return generateEnvChecks(".env.template")
	}

	return nil
}
