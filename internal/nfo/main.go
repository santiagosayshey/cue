package nfo

import (
	"encoding/xml"
	"os"
)

type UniqueID struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

type Info struct {
	Title     string     `xml:"title"`
	UniqueIDs []UniqueID `xml:"uniqueid"`
}

func Parse(path string, target any) error {
	nfoData, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return xml.Unmarshal(nfoData, target)
}
