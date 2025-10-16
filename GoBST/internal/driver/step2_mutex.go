package driver

import (
	"gobst/internal/bst"
	"sync"
)

func Step2Mutexes(trees []*bst.Tree, hashWorkers int) (map[int][]int, []int){
	//implementation B

	n := len(trees)
	jobs := make(chan int, n)
	hashes := make([]int, n)
	buckets := make(map[int][]int) //return

	var wg sync.WaitGroup
	var lock sync.Mutex

	for h := 0; h < hashWorkers; h++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range jobs {
				h := trees[t].HashValue()
				lock.Lock()
				buckets[h] = append(buckets[h], t)
				lock.Unlock()
			}
		
		}()
	}
	for i := 0; i < n; i++ {
		jobs <- i
	}
	wg.Wait()
	close(jobs)

	return buckets, hashes
}