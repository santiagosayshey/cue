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
	"github.com/santiagosayshey/cue/internal/stats"
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
	concurrency := flag.Int("concurrency", 3, "maximum concurrent downloads")
	flag.Parse()
	var st stats.Stats

	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	db, err := database.Load(cfg.Database)
	if err != nil {
		return fmt.Errorf("loading database: %w", err)
	}

	downloaders := downloader.NewDownloaders()
	var jobs []downloader.Job

	for _, library := range cfg.Library {
		st.IncrementLibraries()
		folders, err := filesystem.GetFolders(library.Path)
		if err != nil {
			return fmt.Errorf("reading library folder %s: %w", library.Path, err)
		}

		for _, folderPath := range folders {
			st.IncrementFolders()
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
				st.IncrementMatched()

				for _, asset := range entry.Assets {
					sourceDownloader, ok := downloaders[asset.Source]
					if !ok {
						fmt.Printf("Unknown source %q for %s, skipping\n", asset.Source, info.Title)
						continue
					}
					st.IncrementDownloads()
					jobs = append(jobs, downloader.Job{
						Downloader: sourceDownloader,
						URL:        asset.URL,
						DestPath:   folderPath,
						Filename:   asset.Filename,
						Title:      info.Title,
					})
				}
			}
		}
	}
	fmt.Printf("Scanned %v libraries... %v folders, %v matched\n", st.Libraries, st.Folders, st.Matched)
	fmt.Printf("Downloading %v asset(s) (concurrency %v)...\n", st.Downloads, *concurrency)
	pool := downloader.NewPool(*concurrency, func(title, filename string, err error) {
		if err != nil {
			fmt.Fprintf(os.Stderr, "[error] %s - %s: %v\n", title, filename, err)
			return
		}
		fmt.Printf("[success] %s - %s\n", title, filename)
	})
	for _, job := range jobs {
		pool.Queue(job)
	}
	pool.Wait()
	return nil
}
