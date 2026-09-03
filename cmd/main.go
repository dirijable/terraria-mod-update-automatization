package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/dirijable/terraria-mod-update-automatization/internal/app"
	"github.com/dirijable/terraria-mod-update-automatization/internal/git"
)

func main() {
	pathToSteamTModContent := flag.String("sp", "", "path to steam workshop tmodloader content")
	pathToRepo := flag.String("pp", "", "path to server git repo dir")
	flag.Parse()

	if *pathToSteamTModContent == "" || *pathToRepo == "" {
		fmt.Println("Usage: program -sp [path_to_steam_content] -pp [path_to_server_git_dir]")
		os.Exit(1)
	}

	cleanedSteamPath := filepath.Clean(*pathToSteamTModContent)
	repoPath := *pathToRepo

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		log.Println("\nInterrupt signal received. Rolling back server repository...")
		if err := git.ResetHard(repoPath); err != nil {
			log.Printf("Unable to hard reset: %v", err)
		}
		os.Exit(1)
	}()

	if err := app.Run(ctx, cleanedSteamPath, repoPath); err != nil {
		log.Fatalf("Execution error: %v", err)
	}
}