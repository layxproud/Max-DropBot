package storage

import "sync"

type Deduplicator struct {
	m sync.Map
}

func NewDeduplicator() *Deduplicator {
	return &Deduplicator{}
}

func (d *Deduplicator) Seen(hash string) bool {
	_, ok := d.m.Load(hash)
	return ok
}

func (d *Deduplicator) Store(hash string) {
	d.m.Store(hash, struct{}{})
}
