package downloader

import (
	"sync"
)

type Pool struct {
	wg     sync.WaitGroup
	sem    chan struct{}
	onDone func(title, filename string, err error)
}

type Job struct {
	Downloader Downloader
	URL        string
	DestPath   string
	Filename   string
	Title      string
}

func NewPool(max int, onDone func(title, filename string, err error)) *Pool {
	return &Pool{
		sem:    make(chan struct{}, max),
		onDone: onDone,
	}
}

func (p *Pool) Queue(job Job) {
	p.sem <- struct{}{}
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		defer func() { <-p.sem }()
		err := job.Downloader.Download(job.URL, job.DestPath, job.Filename)
		p.onDone(job.Title, job.Filename, err)
	}()
}

func (p *Pool) Wait() {
	p.wg.Wait()
}
