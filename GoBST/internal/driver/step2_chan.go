//implementation A
package driver

import (
	"sync"
	"gobst/internal/bst"
)
type hashRes struct {
	id, h int
}

func Step2Chan(trees []*bst.Tree, hashWorkers int) (map[int][]int, []int) {
	n := len(trees)
	out := make(chan hashRes, n)
	jobs  := make(chan int, n)
	hashes := make([]int, n)
	buckets := make(map[int][]int) //return

	var wg sync.WaitGroup
	for h := 0; h < hashWorkers; h++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs{
				out <- hashRes{id: i, h: trees[i].HashValue()}
			}
		}()
	}
	for i := 0; i < n; i++ {
		jobs <- i
	}
	close(jobs);

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