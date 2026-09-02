package app

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"strings"

	"terraria-mod-update-automatization/internal/cache"
	"terraria-mod-update-automatization/internal/git"
	"terraria-mod-update-automatization/internal/mods"
	"terraria-mod-update-automatization/internal/steam"
)

const (
	steamAPIURL   = "https://api.steampowered.com/ISteamRemoteStorage/GetPublishedFileDetails/v1/"
	cacheFileName = "mods_cache_dirijable.json"
)

func Run(ctx context.Context, steamPath, repoPath string) error {
	localTmodPaths := findLocalMods(steamPath)
	if len(localTmodPaths) == 0 {
		log.Println("No valid tmod files found on disk.")
		return nil
	}

	modIDs := mapKeysToSlice(localTmodPaths)
	steamResp, err := steam.FetchPublishedFileDetails(ctx, steamAPIURL, modIDs)
	if err != nil {
		return err
	}

	cacheFilePath := cache.GetCachePath(cacheFileName)
	modsToUpdate, updatedCache, err := checkUpdates(cacheFilePath, localTmodPaths, steamResp)
	if err != nil {
		return err
	}

	if len(modsToUpdate) == 0 {
		log.Println("All mods are up to date.")
		return nil
	}

	log.Printf("Found %d updated mod(s). Starting deploy...", len(modsToUpdate))
	return deployMods(repoPath, cacheFilePath, modsToUpdate, updatedCache)
}

func findLocalMods(steamPath string) map[string]string {
	filteredDirs := mods.MustFilteredDirsFromPath(steamPath)
	return mods.LatestTmodFiles(steamPath, filteredDirs)
}

func checkUpdates(cacheFilePath string, localTmodPaths map[string]string, steamResp *steam.SteamResponse) (map[string]string, map[string]int64, error) {
	cch, err := cache.LoadCacheFromFile(cacheFilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load cache: %w", err)
	}

	modsToUpdate := make(map[string]string)
	var newCount, updateCount int

	for _, detail := range steamResp.Response.FileDetails {
		if detail.Result != 1 {
			log.Printf("mod id=%s skipped (banned/deleted, result=%d)", detail.PublishedFileId, detail.Result)
			continue
		}

		srcPath, ok := localTmodPaths[detail.PublishedFileId]
		if !ok {
			continue
		}

		modFileName := filepath.Base(srcPath)
		cachedTime, exists := cch[detail.PublishedFileId]

		if !exists || cachedTime == 0 {
			log.Printf("[NEW] Found new mod: %s (ID: %s)", modFileName, detail.PublishedFileId)
			modsToUpdate[detail.PublishedFileId] = srcPath
			cch[detail.PublishedFileId] = detail.TimeUpdated
			newCount++
			continue
		}

		if detail.TimeUpdated > cachedTime {
			log.Printf("[UPDATE] Found update for: %s (ID: %s)", modFileName, detail.PublishedFileId)
			modsToUpdate[detail.PublishedFileId] = srcPath
			cch[detail.PublishedFileId] = detail.TimeUpdated
			updateCount++
		}
	}

	if len(modsToUpdate) > 0 {
		log.Printf("Total queued for deploy: %d (New: %d, Updates: %d)", len(modsToUpdate), newCount, updateCount)
	}

	return modsToUpdate, cch, nil
}

func deployMods(repoPath, cacheFilePath string, modsToUpdate map[string]string, updatedCache map[string]int64) error {
	if err := git.SyncRepo(repoPath); err != nil {
		return fmt.Errorf("git sync failed: %w", err)
	}

	postavkaDir := filepath.Join(repoPath, "Postavka")
	if err := mods.CopyUpdatedMods(postavkaDir, modsToUpdate); err != nil {
		return fmt.Errorf("copy updated mods failed: %w", err)
	}

	if err := cache.SaveCacheToFile(cacheFilePath, updatedCache); err != nil {
		return fmt.Errorf("failed to save cache: %w", err)
	}

	commitMsg := generateCommitMsg(modsToUpdate)

	if err := git.PushUpdates(repoPath, commitMsg); err != nil {
		return fmt.Errorf("git push failed: %w", err)
	}

	log.Println("Successfully updated and pushed mods!")
	return nil
}

func generateCommitMsg(modsToUpdate map[string]string) string {
	names := make([]string, 0, len(modsToUpdate))
	for _, srcPath := range modsToUpdate {
		filename := filepath.Base(srcPath)
		modName := strings.TrimSuffix(filename, filepath.Ext(filename))
		names = append(names, modName)
	}

	sort.Strings(names)

	return "update mods: " + strings.Join(names, ", ")
}

func mapKeysToSlice(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}
