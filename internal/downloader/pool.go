package downloader

import (
	"fmt"
	"sync"
)

type Pool struct {
	wg  sync.WaitGroup
	sem chan struct{}
}

func NewPool(max int) *Pool {
	return &Pool{
		sem: make(chan struct{}, max),
	}
}

func (p *Pool) Queue(d Downloader, url string, destPath string, filename string, title string) {
	p.sem <- struct{}{}
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		defer func() { <-p.sem }()
		if err := d.Download(url, destPath, filename); err != nil {
			fmt.Printf("Download failed for %s (%s): %v\n", title, filename, err)
			return
		}
		fmt.Printf("Download completed for %s (%s)\n", title, filename)
	}()
}

func (p *Pool) Wait() {
	p.wg.Wait()
}
