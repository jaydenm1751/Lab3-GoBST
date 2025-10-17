package driver

import (
	"sync"
	"gobst/internal/bst"
)

type buffer struct {
	lock sync.Mutex
	NotFull sync.Cond
	NotEmpty sync.Cond
	data []pair
	c int
	closed bool
}

func newBuffer(compWorkers int) *buffer {
	b := &buffer{c: compWorkers}
	b.NotFull = *sync.NewCond(&b.lock)
	b.NotEmpty = *sync.NewCond(&b.lock)
	return b
}

func (b *buffer) get() (pair, bool) {
	b.lock.Lock()
	for len(b.data) == 0 && !b.closed {
		b.NotEmpty.Wait()
	}
	if len(b.data) == 0 && b.closed {
		b.lock.Unlock()
		return pair{}, false
	}

	p := b.data[0]
	copy(b.data[0:], b.data[1:])
	b.data = b.data[:len(b.data) - 1]

	b.NotFull.Signal()
	b.lock.Unlock()

	return p, true

}

func(b *buffer) put(p pair) {
	b.lock.Lock()
	for len(b.data) == b.c && !b.closed {
		b.NotFull.Wait()
	}
	if b.closed {
		b.lock.Unlock()
		return
	}
	b.data = append(b.data, p)
	b.NotEmpty.Signal()
	b.lock.Unlock()
}

func (b *buffer) close() {
	b.lock.Lock()
	b.closed = true
	b.NotEmpty.Broadcast()
	b.NotFull.Broadcast()
	b.lock.Unlock()
}

func Step3Workers(trees []*bst.Tree, buckets map[int][]int, compWorkers int) [][]bool {
	n := len(trees)
	adj := MakeAdj(n)

	t_map := TreesIndex(trees)
	queue := newBuffer(compWorkers)
	var wg sync.WaitGroup
	var lock sync.Mutex

	for w := 0; w < compWorkers; w++ {
		wg.Add(1)
		go func(){
			defer wg.Done()
			for{
				p, proceed := queue.get()
				if !proceed { return }
				if t_map[p.i].Equal(t_map[p.j]){
					lock.Lock()
					adj[p.i][p.j], adj[p.j][p.i] = true, true
					lock.Unlock()
				}
			}
		}()
	}

	for _, ids := range buckets {
		if len(ids) < 2 { continue }
		for _, p := range BuildPairs(ids) {
			queue.put(p)
		}
	}

	queue.close()
	wg.Wait()
	return adj
}
