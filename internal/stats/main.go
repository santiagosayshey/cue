package stats

import "sync"

type Stats struct {
	mu        sync.Mutex
	Libraries int
	Folders   int
	Matched   int
	Downloads int
	Failed    int
}

func (s *Stats) IncrementLibraries() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Libraries++
}

func (s *Stats) IncrementFolders() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Folders++
}

func (s *Stats) IncrementMatched() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Matched++
}

func (s *Stats) IncrementDownloads() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Downloads++
}
