package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// D:\SSSteam\steamapps\workshop\content\1281930
// C:\Users\user\Space\DevOps\terka\terraria-server\terraria-server

const (
	steamAPIURL = "https://api.steampowered.com/ISteamRemoteStorage/GetPublishedFileDetails/v1/"
	cacheFile   = "mods_cache_dirijable.json"
)

func main() {
	pathToSteamTModontent := flag.String("sp", "", "path to steam workshop tmodloader content")
	pathToRepo := flag.String("pp", "", "path to server git repo dir") 
	flag.Parse()
	if *pathToSteamTModontent == "" {
		fmt.Println("Usage: program -sp [path_to_steam_content]")
		os.Exit(1)
	}

	if *pathToRepo == "" {
		fmt.Println("Usage: program -pp [path_to_server_git_dir]")
		os.Exit(1)
	}

	cleanedPath := filepath.Clean(*pathToSteamTModontent)

	filteredDirs := MustFilteredDirsFromPath(cleanedPath)
	localTmodPaths := LatestTmodFiles(cleanedPath, filteredDirs)
	if len(localTmodPaths) == 0 {
		log.Println("No valid tmod files found on disk.")
		return
	}

	formData := SetUrlValues(MapToSliceOfKeys(localTmodPaths))

	resp, err := http.PostForm(steamAPIURL, formData)
	if err != nil {
		log.Fatalf("post mod ids to steam: %v", err)
	}
	defer resp.Body.Close()

	var steamResp SteamResponse
	if err := json.NewDecoder(resp.Body).Decode(&steamResp); err != nil {
		log.Fatalf("parse steam json response: %v", err)
	}

	cache, err := LoadCacheFromFile(cacheFile)
	if err != nil {
		log.Fatalf("failed to load cache: %v", err)
	}

	modsToUpdate := make(map[string]string)
	for _, detail := range steamResp.Response.FileDetails {
		if detail.Result != 1 {
			log.Printf("mod id=%s skipped (banned/deleted, result=%d)", detail.PublishedFileId, detail.Result)
			continue
		}

		cachedTime := cache[detail.PublishedFileId]
		if detail.TimeUpdated > cachedTime {
			srcPath, ok := localTmodPaths[detail.PublishedFileId]
			if !ok {
				continue
			}

			modsToUpdate[detail.PublishedFileId] = srcPath
			cache[detail.PublishedFileId] = detail.TimeUpdated
		}
	}

	if len(modsToUpdate) == 0 {
		log.Println("All mods are up to date.")
		return
	}	

	log.Printf("Found %d updated mod(s). Starting deploy...", len(modsToUpdate))

	if err := SyncRepo(*pathToRepo); err != nil {
		log.Fatalf("git sync failed: %v", err)
	}

	if err := CopyUpdatedMods(filepath.Join(*pathToRepo, "Postavka"), modsToUpdate); err != nil {
		log.Fatalf("copy updated mods failed: %v", err)
	}

	if err := SaveCacheToFile(cacheFile, cache); err != nil {
		log.Fatalf("failed to save cache: %v", err)
	}

	if err := PushUpdates(*pathToRepo, "auto-update tModLoader mods"); err != nil {
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
