package pkg

import (
	"os"
)

func deducePathFromEnv(path string) string {
	if _, ok := os.LookupEnv("TEST"); ok {
		return path[4:]
	}

	return path
}

func ReadMarkdownFile(path string) ([]byte, error) {
	content, err := os.ReadFile(deducePathFromEnv(path))
	if err != nil {
		return nil, err
	}

	return content, nil
}
