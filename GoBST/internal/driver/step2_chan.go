//implementation A
package driver

import (
	"sync"
	"gobst/internal/bst"
)
type hashRes struct {
	id, h int
}

func Step2Chan(trees []*bst.Tree) (map[int][]int, []int) {
	n := len(trees)
	out := make(chan hashRes, n)
	hashes := make([]int, n)
	buckets := make(map[int][]int) //return

	var wg sync.WaitGroup
	for i, t := range trees {
		wg.Add(1)
		go func(i int, t *bst.Tree) {
			defer wg.Done()
			out <- hashRes{id: i, h: t.HashValue()}
		}(i, t)
	}

	go func() {	
		wg.Wait()
		close(out)
	}()

	for r := range out {
		hashes[r.id] = r.h
		buckets[r.h] = append(buckets[r.h], r.id)
	}
	return buckets, hashes
}