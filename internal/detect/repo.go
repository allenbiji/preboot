package detect

import "github.com/allenbiji/clone-sage/internal/model"

func ScanRepo() []model.CheckConfig {
	var checks []model.CheckConfig

	checks = append(checks, detectGo()...)
	checks = append(checks, detectEnv()...)

	if fileExists("Makefile") {
		checks = append(checks, model.CheckConfig{
			Name:     "make-installed",
			Type:     model.TypeCommandExists,
			Severity: model.SeverityBlocker,
			Options:  map[string]string{"command": "make"},
			Message:  "Makefile has been detected. 'make' is required for build tasks",
		})
	}

	return checks

}
