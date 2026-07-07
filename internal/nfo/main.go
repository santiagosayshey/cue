package nfo

import (
	"encoding/xml"
	"os"
	"strings"
)

type UniqueID struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

type NFO struct {
	Title     string     `xml:"title"`
	UniqueIDs []UniqueID `xml:"uniqueid"`
}

func Parse(path string) (NFO, error) {
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

func (n NFO) UIDMap() map[string]string {
	ids := make(map[string]string, len(n.UniqueIDs))

	for _, uid := range n.UniqueIDs {
		if uid.Type == "" || uid.Value == "" {
			continue
		}

		ids[uid.Type] = strings.TrimSpace(uid.Value)
	}

	return ids
}
