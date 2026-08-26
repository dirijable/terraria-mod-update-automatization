package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"terraria-mod-update-automatization/internal/cache"
	"terraria-mod-update-automatization/internal/git"
	"terraria-mod-update-automatization/internal/mods"
	"terraria-mod-update-automatization/internal/steam"
)

const (
	steamAPIURL = "https://api.steampowered.com/ISteamRemoteStorage/GetPublishedFileDetails/v1/"
	cacheFile   = "mods_cache_dirijable.json"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pathToSteamTModContent := flag.String("sp", "", "path to steam workshop tmodloader content")
	pathToRepo := flag.String("pp", "", "path to server git repo dir")
	flag.Parse()

	if *pathToSteamTModContent == "" || *pathToRepo == "" {
		fmt.Println("Usage: program -sp [path_to_steam_content] -pp [path_to_server_git_dir]")
		os.Exit(1)
	}

	cleanedSteamPath := filepath.Clean(*pathToSteamTModContent)
	repoPath := *pathToRepo

	go func() {
		<-ctx.Done()
		log.Println("\nInterrupt signal recieved...")
		if err := git.ResetHard(repoPath); err != nil {
			log.Printf("Unable to hard reset: %v", err)
		}
		os.Exit(1)
	}()

	filteredDirs := mods.MustFilteredDirsFromPath(cleanedSteamPath)
	localTmodPaths := mods.LatestTmodFiles(cleanedSteamPath, filteredDirs)
	if len(localTmodPaths) == 0 {
		log.Println("No valid tmod files found on disk.")
		return
	}

	formData := steam.SetUrlValues(MapToSliceOfKeys(localTmodPaths))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, steamAPIURL, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		log.Fatalf("create steam request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("post mod ids to steam: %v", err)
	}
	defer resp.Body.Close()

	var steamResp steam.SteamResponse
	if err := json.NewDecoder(resp.Body).Decode(&steamResp); err != nil {
		log.Fatalf("parse steam json response: %v", err)
	}

	cch, err := cache.LoadCacheFromFile(cacheFile)
	if err != nil {
		log.Fatalf("failed to load cache: %v", err)
	}

	modsToUpdate := make(map[string]string)
	for _, detail := range steamResp.Response.FileDetails {
		if detail.Result != 1 {
			log.Printf("mod id=%s skipped (banned/deleted, result=%d)", detail.PublishedFileId, detail.Result)
			continue
		}

		cachedTime := cch[detail.PublishedFileId]
		if detail.TimeUpdated > cachedTime {
			srcPath, ok := localTmodPaths[detail.PublishedFileId]
			if !ok {
				continue
			}
			modsToUpdate[detail.PublishedFileId] = srcPath
			cch[detail.PublishedFileId] = detail.TimeUpdated
		}
	}

	if len(modsToUpdate) == 0 {
		log.Println("All mods are up to date.")
		return
	}

	log.Printf("Found %d updated mod(s). Starting deploy...", len(modsToUpdate))

	if err := git.SyncRepo(repoPath); err != nil {
		log.Fatalf("git sync failed: %v", err)
	}

	postavkaDir := filepath.Join(repoPath, "Postavka")
	if err := mods.CopyUpdatedMods(postavkaDir, modsToUpdate); err != nil {
		log.Fatalf("copy updated mods failed: %v", err)
	}

	if err := cache.SaveCacheToFile(cacheFile, cch); err != nil {
		log.Fatalf("failed to save cache: %v", err)
	}

	if err := git.PushUpdates(repoPath, "auto-update tModLoader mods"); err != nil {
		log.Fatalf("git push failed: %v", err)
	}

	log.Println("Successfully updated and pushed mods!")
}

func MapToSliceOfKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}