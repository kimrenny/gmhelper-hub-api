package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// loadDotEnv attempts to find and load a .env file from the current directory or any parent directory.
// It will not overwrite variables that are already set in the process environment.
func loadDotEnv(filenames ...string) {
	if len(filenames) > 0 {
		for _, filename := range filenames {
			if loadFile(filename) {
				return
			}
		}
		return
	}

	// Default: check current directory and walk up parent directories
	curr, err := os.Getwd()
	if err != nil {
		curr = "."
	}

	for {
		target := filepath.Join(curr, ".env")
		if loadFile(target) {
			return
		}

		parent := filepath.Dir(curr)
		if parent == curr || parent == "" {
			break
		}
		curr = parent
	}
}

func loadFile(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		// Strip surrounding quotes if present
		if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
			val = val[1 : len(val)-1]
		}

		// Only set if not already set in process environment
		if os.Getenv(key) == "" && key != "" {
			_ = os.Setenv(key, val)
		}
	}
	return true
}
