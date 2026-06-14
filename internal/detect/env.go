package detect

import (
	"strings"

	"github.com/allenbiji/clone-sage/internal/model"
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

	keys := extractEnvKeys(fileName)

	for _, key := range keys {
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
