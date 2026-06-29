package detect

import (
	"github.com/allenbiji/preboot/internal/model"
)

func detectGo() []model.CheckConfig {
	var checks []model.CheckConfig

	if fileExists("go.mod") || fileExists("go.work") {
		checks = append(checks, model.CheckConfig{
			Name:     "go-installed",
			Type:     model.TypeCommandExists,
			Severity: model.SeverityBlocker,
			Options:  map[string]string{"command": "go"},
			Message:  "The project does not have go installed",
		})
	}

	return checks
}
