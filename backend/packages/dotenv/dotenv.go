package dotenv

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

/**
 * Loads key-value pairs from files and puts them on the runtime environment
 * If you call dotenv.Load() without arguments, it will read all files starting with .env from the current working directory
 * But also accepts a custom glob pattern, such as dotenv.Laod(".env.local")
 */
func Load(globPattern ...string) error {
	pattern := ".env*"
	if len(globPattern) == 1 {
		pattern = globPattern[0]
	} else if len(globPattern) > 1 {
		return fmt.Errorf("too many glob pattern arguments: %v ", globPattern)
	}

	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("error finding .env files: %w", err)
	}

	for _, file := range files {
		if err := loadFile(file); err != nil {
			return fmt.Errorf("error loading file %s: %w", file, err)
		}
	}

	return nil
}

func loadFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
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
			return fmt.Errorf("invalid line format: %s", line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("error setting environment variable %s: %w", key, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error scanning file: %w", err)
	}

	return nil
}
