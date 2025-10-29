package utils

import (
	"log"
	"os"
)

func GetFiles(dirPath string) []os.DirEntry {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		log.Fatalf("Failed to get directory path: %v", err)
	}
	return files
}
