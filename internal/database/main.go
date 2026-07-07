package database

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Asset struct {
	Source   string `yaml:"source"`
	URL      string `yaml:"url"`
	Filename string `yaml:"filename"`
}

type Entry struct {
	Title  string  `yaml:"title"`
	Assets []Asset `yaml:"assets"`
}

func readPath(path string) ([]byte, error) {
	if strings.HasPrefix(path, "http") {
		resp, err := http.Get(path)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("fetching %s: %s", path, resp.Status)
		}
		return io.ReadAll(resp.Body)
	}
	return os.ReadFile(path)
}

func Load(databases []string) (map[string]Entry, error) {
	var db map[string]Entry
	for _, dbPath := range databases {
		dbData, err := readPath(dbPath)
		if err != nil {
			return nil, err
		}
		// same keys get overwritten every loop on purpose
		if err := yaml.Unmarshal(dbData, &db); err != nil {
			return nil, err
		}
	}
	return db, nil
}

func Lookup(uids map[string]string, db map[string]Entry) (Entry, bool) {
	tmdbID, ok := uids["tmdb"]
	if !ok {
		return Entry{}, false
	}

	item, found := db[tmdbID]
	return item, found
}
