package driver

import (
	"gobst/internal/bst"
	"sync"
	// "fmt"
	"time"
)

func Step2Mutexes(trees []*bst.Tree, hashWorkers int) (map[int][]int, time.Duration, time.Duration){
	//implementation B

	n := len(trees)
	jobs := make(chan int, n)
	//hashes := make([]int, n)
	buckets := make(map[int][]int) //return

	var wgAll sync.WaitGroup
	var wg sync.WaitGroup
	var lock sync.Mutex
	wgAll.Add(hashWorkers)
	wg.Add(n)

	start := time.Now()

	// fmt.Printf("reaching the for loop\n")

	for h := 0; h < hashWorkers; h++ {
		//wg.Add(1)
		go func() {
			defer wgAll.Done()
			for t := range jobs {
				h := trees[t].HashValue()
				wg.Done()
				lock.Lock()
				//hashes[t] = h
				buckets[h] = append(buckets[h], t)
				lock.Unlock()
			}
		
		}()
	}
	for i := 0; i < n; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	hashTime := time.Since(start)

	wgAll.Wait()
	hashGroupTime := time.Since(start)

	return buckets, hashTime, hashGroupTime
}