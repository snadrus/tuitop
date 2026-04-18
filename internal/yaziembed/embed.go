package yaziembed

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed all:assets/yazi
var embedded embed.FS

const bundleKey = "v4"

// ConfigDir returns a filesystem directory containing the embedded Yazi config tree.
// Files are written under the user cache dir and reused when the bundle key matches.
func ConfigDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dest := filepath.Join(base, "tuitop", "yazi-bundled", bundleKey)
	marker := filepath.Join(dest, ".tuitop-bundle")
	if b, err := os.ReadFile(marker); err == nil && strings.TrimSpace(string(b)) == bundleKey {
		return dest, nil
	}
	if err := os.RemoveAll(dest); err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return "", err
	}
	const root = "assets/yazi"
	err = fs.WalkDir(embedded, root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(path, root), "/")
		if rel == "" {
			return nil
		}
		out := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(out, 0o755)
		}
		data, err := embedded.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		return os.WriteFile(out, data, 0o644)
	})
	if err != nil {
		return "", fmt.Errorf("yaziembed: extract: %w", err)
	}
	if err := os.WriteFile(marker, []byte(bundleKey), 0o644); err != nil {
		return "", err
	}
	return dest, nil
}
