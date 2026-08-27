package cache

import (
	"os"
	"path/filepath"
)

func GetCachePath(filename string) string {
	baseCacheDir, err := os.UserCacheDir()
	if err != nil {
		return filename
	}

	appCacheDir := filepath.Join(baseCacheDir, "terraria-mod-sync")
	if err := os.MkdirAll(appCacheDir, 0755); err != nil {
		return filename
	}

	return filepath.Join(appCacheDir, filename)
}