package assets

import (
	"os"
	"path/filepath"
)

func resolvePalettePath(basePath string) string {
	candidate := basePath + ".palette"
	if fileExists(candidate) {
		return candidate
	}
	dir := filepath.Dir(basePath)
	for {
		candidate = filepath.Join(dir, "default.palette")
		if fileExists(candidate) {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return candidate
		}
		dir = parent
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
