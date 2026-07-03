package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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

type Theme struct {
	Source string `json:"source"`
	URL    string `json:"url"`
}

type Entry struct {
	Title string `json:"title"`
	Theme Theme  `json:"theme"`
}

type UniqueID struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

type NFO struct {
	Title     string     `xml:"title"`
	UniqueIDs []UniqueID `xml:"uniqueid"`
}

func main() {
	// read in the config file
	cfgPath := os.Args[1]
	cfgData, err := os.ReadFile(cfgPath)
	if err != nil {
		fmt.Println("Error reading config:", err)
		return
	}
	var cfg Config
	err = yaml.Unmarshal(cfgData, &cfg)
	if err != nil {
		fmt.Println("Error parsing config:", err)
		return
	}

	// build database, later keys win
	database := map[string]Entry{}
	for _, dbPath := range cfg.Database {
		dbData, err := os.ReadFile(dbPath)
		if err != nil {
			fmt.Println("Error reading database:", err)
			return
		}
		err = json.Unmarshal(dbData, &database)
		if err != nil {
			fmt.Println("Error parsing database:", err)
			return
		}
	}

	// iterate over each library, then each entry inside a library

	for _, library := range cfg.Library {
		fmt.Println(library.Path)
		items, err := os.ReadDir(library.Path)
		if err != nil {
			fmt.Println("Error reading library folder:", err)
			return
		}

		for _, i := range items {
			fmt.Println("- " + i.Name())
			itemPath := filepath.Join(library.Path, i.Name())
			itemFiles, err := os.ReadDir(itemPath)
			if err != nil {
				fmt.Printf("Error reading %v:", i.Name())
				return
			}

			for _, f := range itemFiles {
				fmt.Println("-- " + f.Name())

				if filepath.Ext(f.Name()) == ".nfo" {
					nfoPath := filepath.Join(itemPath, f.Name())
					nfoData, err := os.ReadFile(nfoPath)
					if err != nil {
						fmt.Println("Error reading NFO:", err)
						return
					}

					var nfo NFO
					err = xml.Unmarshal(nfoData, &nfo)
					if err != nil {
						fmt.Println("Error parsing NFO:", err)
						return
					}

					var entry Entry
					for _, id := range nfo.UniqueIDs {
						if id.Type == "tmdb" {
							entry = database[id.Value]
						}
					}

					if entry.Theme.URL == "" {
						fmt.Printf("No theme found for %s, skipping\n", nfo.Title)
						continue
					}

					cmd := exec.Command("yt-dlp", "-x", "--audio-format", "mp3", "-o", "theme.mp3", entry.Theme.URL)
					cmd.Dir = itemPath
					err = cmd.Run()
				}
			}
		}
	}
}
