package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/santiagosayshey/cue/internal/config"
	"github.com/santiagosayshey/cue/internal/database"
	"github.com/santiagosayshey/cue/internal/downloader"
	"github.com/santiagosayshey/cue/internal/filesystem"
	"github.com/santiagosayshey/cue/internal/nfo"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	configPath := flag.String("config", "config.yaml", "path to config file")
	concurrency := flag.Int("concurrency ", 10, "maximum concurrent downloads")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	db, err := database.Load(cfg.Database)
	if err != nil {
		return fmt.Errorf("loading database: %w", err)
	}

	downloaders := downloader.NewDownloaders()
	pool := downloader.NewPool(*concurrency)

	for _, library := range cfg.Library {
		folders, err := filesystem.GetFolders(library.Path)
		if err != nil {
			return fmt.Errorf("reading library folder %s: %w", library.Path, err)
		}

		for _, folderPath := range folders {
			nfoPath, err := filesystem.GetFile(folderPath, ".nfo")
			if err != nil {
				fmt.Println("Skipping folder, no NFO found:", folderPath)
				continue
			}

			var info nfo.Info
			if err := nfo.Parse(nfoPath, &info); err != nil {
				fmt.Printf("Couldn't load NFO: %v\n", nfoPath)
				continue
			}

			for _, uid := range info.UniqueIDs {
				entry, found := db[uid.Value]
				if !found {
					continue
				}

				for _, asset := range entry.Assets {
					sourceDownloader, ok := downloaders[asset.Source]
					if !ok {
						fmt.Printf("Unknown source %q for %s, skipping\n", asset.Source, info.Title)
						continue
					}
					pool.Queue(sourceDownloader, asset.URL, folderPath, asset.Filename, info.Title)
				}
			}
		}
	}

	pool.Wait()
	return nil
}
