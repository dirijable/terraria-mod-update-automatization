package mods

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const dirVersionLayout = "2006.1"

func LatestTmodFiles(basePath string, filteredTmodDirs []os.DirEntry) map[string]string {
	latestFiles := make(map[string]string, len(filteredTmodDirs))

	for _, dir := range filteredTmodDirs {
		modID := dir.Name()
		modPath := filepath.Join(basePath, modID)

		filteredVersionDirs, err := FilteredDirsFromPath(modPath)
		if err != nil {
			log.Printf("skip mod id=%q: %v", modID, err)
			continue
		}

		latestFile, err := latestTmodFile(modPath, filteredVersionDirs)
		if err != nil {
			log.Printf("mod id=%q will not update: %v", filepath.Base(modPath), err)
			continue
		}

		latestFiles[modID] = latestFile
	}

	return latestFiles
}

func latestTmodFile(modPath string, filteredVersionDirs []os.DirEntry) (string, error) {
	var (
		latestDir  os.DirEntry
		latestTime time.Time
	)

	for _, dir := range filteredVersionDirs {
		t, err := time.Parse(dirVersionLayout, dir.Name())
		if err != nil {
			continue
		}
		if t.After(latestTime) {
			latestTime = t
			latestDir = dir
		}
	}
	if latestDir == nil {
		return "", fmt.Errorf("no valid version folders found in %s", modPath)
	}

	targetPath := filepath.Join(modPath, latestDir.Name())
	files, err := os.ReadDir(targetPath)
	if err != nil {
		return "", fmt.Errorf("read target dir %s: %w", targetPath, err)
	}

	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".tmod") {
			return filepath.Join(targetPath, f.Name()), nil
		}
	}

	return "", fmt.Errorf("no .tmod file found in %s", targetPath)
}
