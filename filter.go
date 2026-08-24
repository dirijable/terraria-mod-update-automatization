package main

import "os"

var OnlyDirPredicate = func(dir os.DirEntry) bool {
	return dir.IsDir()
}

func Filter[T any](entities []T, predicate func(e T) bool) []T {
	result := make([]T, 0)
	for _, e := range entities {
		if predicate(e) {
			result = append(result, e)
		}
	}
	return result
}
