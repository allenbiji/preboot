package detect

import (
	"github.com/allenbiji/clone-sage/internal/model"
)

func detectGo() []model.CheckConfig {
	var checks []model.CheckConfig

	if fileExists("go.mod") {
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
