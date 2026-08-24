package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

type Cache map[string]int64

func NewCache() Cache {
	return make(Cache)
}

func Decode(r io.Reader) (Cache, error) {
	cache := make(Cache)
	err := json.NewDecoder(r).Decode(&cache)
	return cache, err
}

func Encode(w io.Writer, cache Cache) error {
	return json.NewEncoder(w).Encode(cache)
}

func LoadCacheFromFile(path string) (Cache, error) {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return NewCache(), nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cache, err := Decode(file)
	if err != nil {
		return nil, fmt.Errorf("decode cache file: %w", err)
	}
	return cache, nil
}

func SaveCacheToFile(path string, cache Cache) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return Encode(file, cache)
}
