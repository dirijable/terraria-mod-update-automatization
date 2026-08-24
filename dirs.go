package main

import (
	"log"
	"os"
)

func FilteredDirsFromPath(path string) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	return Filter(entries, OnlyDirPredicate), nil
}

func MustDirsFromPath(path string) []os.DirEntry {
	dirs, err := os.ReadDir(path)
	if err != nil {
		log.Fatalf("get tmod ids: %v", err)
	}
	return dirs
}
