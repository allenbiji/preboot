package detect

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// looks for the filename to verify that it exists
func fileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// extract the env keys from the .env.example file
func ExtractEnvKeys(filePath string) (map[string]string, error) {
	envMap := make(map[string]string)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open %s: %w", filePath, err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if before, _, found := strings.Cut(line, "#"); found {
			line = strings.TrimSpace(before)
		}

		key, value, found := strings.Cut(line, "=")
		if found {
			envMap[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading %s: %w", filePath, err)
	}

	return envMap, nil
}
