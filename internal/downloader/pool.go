package downloader

import (
	"sync"
)

type Pool struct {
	wg     sync.WaitGroup
	sem    chan struct{}
	onDone func(title, filename string, err error)
}

func NewPool(max int, onDone func(title, filename string, err error)) *Pool {
	return &Pool{
		sem:    make(chan struct{}, max),
		onDone: onDone,
	}
}

func (p *Pool) Queue(d Downloader, url string, destPath string, filename string, title string) {
	p.sem <- struct{}{}
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		defer func() { <-p.sem }()
		err := d.Download(url, destPath, filename)
		p.onDone(title, filename, err)
	}()
}

func (p *Pool) Wait() {
	p.wg.Wait()
}
