package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/santiagosayshey/cue/internal/config"
	"github.com/santiagosayshey/cue/internal/database"
	"github.com/santiagosayshey/cue/internal/nfo"
)

func main() {
	downloaders := newDownloaders()

	// flags
	configPath := flag.String("config", "config.yaml", "path to config file")
	down := flag.Int("down", 10, "maximum concurrent downloads")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Println("Couldn't load config: ", err)
		os.Exit(1)
	}

	db, err := database.Load(cfg.Database)
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
					p_nfo, err := nfo.Parse(nfoPath)
					if err != nil {
						fmt.Printf("Couldn't load NFO: %v \n\n %v ", nfoPath, err)
						os.Exit(1)
					}

					entry, found := database.Lookup(p_nfo.UIDMap(), db)
					if !found {
						continue
					}
					for _, asset := range entry.Assets {
						downloader, ok := downloaders[asset.Source]
						if !ok {
							fmt.Printf("Unknown source %q for %s, skipping\n", asset.Source, p_nfo.Title)
							continue
						}
						maxDownloads <- struct{}{}
						wg.Add(1)
						go func() {
							defer wg.Done()
							defer func() { <-maxDownloads }()
							err := downloader.Download(asset.URL, itemPath, asset.Filename)
							if err != nil {
								fmt.Printf("Download failed for %s (%s): %v\n", p_nfo.Title, asset.Filename, err)
								return
							}
							fmt.Printf("Download completed for %s (%s)\n", p_nfo.Title, asset.Filename)
						}()
					}
				}
			}
		}
	}

	wg.Wait()
}
