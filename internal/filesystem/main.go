package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GetFolders(path string) ([]string, error) {
	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var folders []string
	for _, e := range dirEntries {
		if e.IsDir() {
			folders = append(folders, filepath.Join(path, e.Name()))
		}
	}
	return folders, nil
}

func GetFile(folderPath string, desiredFile string) (string, error) {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return "", err
	}
	for _, f := range files {
		if strings.Contains(f.Name(), desiredFile) {
			return filepath.Join(folderPath, f.Name()), nil
		}
	}
	return "", fmt.Errorf("%s not found in %s", desiredFile, folderPath)
}
