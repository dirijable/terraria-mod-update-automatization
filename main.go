package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

//~/Space/DevOps/terka/terraria-server
// D:\SSSteam\steamapps\workshop\content\1281930

type TModDir struct {
	dir      string
	tmodFile string
}

func main() {
	pathToTModSteamContent := flag.String("p", "", "-p [path]")
	flag.Parse()
	if *pathToTModSteamContent == "" {
		fmt.Println("you should give the path")
		os.Exit(1)
	}
	// file, err := os.OpenFile("hashes.txt", os.O_CREATE|os.O_RDWR, 0666)
	// if err != nil {
	// 	fmt.Printf("open file: %v", err)
	// 	os.Exit(1)
	// }
	// defer file.Close()

	wg := sync.WaitGroup{}
	wg.Add(2)
	fileMap := make(map[string]TModDir)
	go func ()  {
		defer wg.Done()
		_ = GetNewest(*pathToTModSteamContent, fileMap)
	}()
		
}

func GetNewest(dir string, fileMap map[string]TModDir) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			path := filepath.Join(dir, e.Name())
			if err := GetNewest(path, fileMap); err != nil {
				return err
			}
		} else {
			if !strings.HasSuffix(e.Name(), ".tmod") {
				continue
			}
			parentDir := filepath.Dir(dir)
			baseDir := filepath.Base(dir)
			if existedDir, ok := fileMap[parentDir]; !ok || isNewer(baseDir, existedDir.dir) {
				fileMap[parentDir] = TModDir{
					dir:      baseDir,
					tmodFile: filepath.Join(dir, e.Name())}
			}
		}
	}
	return nil
}

func isNewer(current, existed string) bool {
	currParts := strings.Split(current, ".")
	exParts := strings.Split(existed, ".")

	if len(currParts) != 2 || len(exParts) != 2 {
		return current > existed 
	}

	currYear, _ := strconv.Atoi(currParts[0])
	currMonth, _ := strconv.Atoi(currParts[1])

	exYear, _ := strconv.Atoi(exParts[0])
	exMonth, _ := strconv.Atoi(exParts[1])

	if currYear != exYear {
		return currYear > exYear
	}
	return currMonth > exMonth
}