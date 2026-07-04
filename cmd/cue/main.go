package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Library  []Library `yaml:"libraries"`
	Database []string  `yaml:"databases"`
}

type Library struct {
	Path string `yaml:"path"`
	Type string `yaml:"type"`
}

type Asset struct {
	Source   string `yaml:"source"`
	URL      string `yaml:"url"`
	Filename string `yaml:"filename"`
}

type MediaItem struct {
	Title  string  `yaml:"title"`
	Assets []Asset `yaml:"assets"`
}

type UniqueID struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

type NFO struct {
	Title     string     `xml:"title"`
	UniqueIDs []UniqueID `xml:"uniqueid"`
}

func loadConfig(path string) (Config, error) {
	cfgData, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	err = yaml.Unmarshal(cfgData, &cfg)
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
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

func loadDatabase(databases []string) (map[string]MediaItem, error) {
	var db map[string]MediaItem
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

func parseNFO(path string) (NFO, error) {
	nfoData, err := os.ReadFile(path)
	if err != nil {
		return NFO{}, err
	}

	var nfo NFO
	err = xml.Unmarshal(nfoData, &nfo)
	if err != nil {
		return NFO{}, err
	}
	return nfo, nil
}

func lookupMediaItem(nfo NFO, database map[string]MediaItem) (MediaItem, bool) {
	for _, id := range nfo.UniqueIDs {
		if id.Type == "tmdb" {
			item, found := database[id.Value]
			return item, found
		}
	}
	return MediaItem{}, false
}

func main() {
	downloaders := newDownloaders()

	// flags
	configPath := flag.String("config", "config.yaml", "path to config file")
	down := flag.Int("down", 10, "maximum concurrent downloads")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Println("Couldn't load config: ", err)
		os.Exit(1)
	}

	database, err := loadDatabase(cfg.Database)
	if err != nil {
		fmt.Println("Couldn't load database: ", err)
		os.Exit(1)
	}

	// concurrent handling
	var wg sync.WaitGroup
	// max
	maxDownloads := make(chan struct{}, *down)

	// library
	for _, library := range cfg.Library {
		items, err := os.ReadDir(library.Path)
		if err != nil {
			fmt.Println("Error reading library folder:", err)
			return
		}

		// library folders
		for _, i := range items {
			itemPath := filepath.Join(library.Path, i.Name())
			itemFiles, err := os.ReadDir(itemPath)
			if err != nil {
				fmt.Printf("Error reading %v:", i.Name())
				return
			}

			// items inside library folders
			for _, f := range itemFiles {

				if filepath.Ext(f.Name()) == ".nfo" {

					nfoPath := filepath.Join(itemPath, f.Name())
					nfo, err := parseNFO(nfoPath)
					if err != nil {
						fmt.Printf("Couldn't load NFO: %v \n\n %v ", nfoPath, err)
						os.Exit(1)
					}

					mediaItem, found := lookupMediaItem(nfo, database)
					if !found {
						// fmt.Printf("No entry for %s, skipping\n", nfo.Title)
						continue
					}

					for _, asset := range mediaItem.Assets {
						downloader, ok := downloaders[asset.Source]
						if !ok {
							fmt.Printf("Unknown source %q for %s, skipping\n", asset.Source, nfo.Title)
							continue
						}
						maxDownloads <- struct{}{}
						wg.Add(1)
						go func() {
							defer wg.Done()
							defer func() { <-maxDownloads }()
							err := downloader.Download(asset.URL, itemPath, asset.Filename)
							if err != nil {
								fmt.Printf("Download failed for %s (%s): %v\n", nfo.Title, asset.Filename, err)
							}
							fmt.Printf("Download completed for %s (%s)", nfo.Title, asset.Filename)
						}()
					}
				}
			}
		}
	}

	wg.Wait()
}
