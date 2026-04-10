package downloader

import (
	"context"
	"sync"
)

type Job struct {
	URL      string
	Filename string
	Path     string
	ChatID   int64
	Ctx      context.Context
}

type Result struct {
	ChatID  int64
	File    string
	Status  string
	Message string
}

type Pool struct {
	jobs    chan Job
	results chan Result
	dl      *Downloader
	wg      sync.WaitGroup
}

func NewPool(workers int) *Pool {
	p := &Pool{
		jobs:    make(chan Job, 100),
		results: make(chan Result, 100),
		dl:      NewDownloader(),
	}

	for i := 0; i < workers; i++ {
		p.wg.Add(1)
		go p.worker()
	}

	return p
}

func (p *Pool) Submit(job Job) {
	p.jobs <- job
}

func (p *Pool) Results() <-chan Result {
	return p.results
}

func (p *Pool) SendResult(res Result) {
	p.results <- res
}

func (p *Pool) Close() {
	close(p.jobs)
	p.wg.Wait()
	close(p.results)
}

func (p *Pool) worker() {
	defer p.wg.Done()
	for job := range p.jobs {
		res := p.dl.Download(job)
		p.results <- res
	}
}
