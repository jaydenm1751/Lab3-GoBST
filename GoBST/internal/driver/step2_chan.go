//implementation A
package driver

import (
	"sync"
	"gobst/internal/bst"
	"time"
)
type hashRes struct {
	id, h int
}

func Step2Chan(trees []*bst.Tree, hashWorkers int) (map[int][]int, time.Duration, time.Duration) {
	n := len(trees)
	out := make(chan hashRes, n)
	jobs  := make(chan int, n)
	// hashes := make([]int, n)

	var wg sync.WaitGroup
	start := time.Now()
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
	close(jobs)

	wg.Wait()
	hashTime := time.Since(start)
	close(out)

	buckets := make(map[int][]int) //return
	for r := range out {
		//hashes[r.id] = r.h
		buckets[r.h] = append(buckets[r.h], r.id)
	}
	hashGroupTime := time.Since(start)
	
	return buckets, hashTime, hashGroupTime
}