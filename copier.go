package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func CopyFile(dst, src string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open src file %s: %w", src, err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create dst file %s: %w", src, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copy data from %s to %s: %w", src, dst, err)
	}

	return nil
}

func CopyUpdatedMods(targetDir string, modsToUpdate map[string]string) error {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("create target dir %s: %w", targetDir, err)
	}

	for _, srcPath := range modsToUpdate {
		fileName := filepath.Base(srcPath)
		dst := filepath.Join(targetDir, fileName)
		if err := CopyFile(dst, srcPath); err != nil {
			return err
		}
	}
	return nil
}