package driver

import (
	//stuff
	"gobst/internal/bst"
	"sync"
)

func buildParallel(lines [][]int, dataWorkers int) []*bst.Tree {
	n := len(lines)

	res := make([]*bst.Tree, n)
	jobs := make(chan int, n)

	var wg sync.WaitGroup
	// for i, vals := range lines {
	// 	t := bst.New()
	// 	for _, v := range vals{
	// 		t.Insert(v)
	// 	}
	// 	trees[i] = t
	// }

	for w := 0; w < dataWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				t := bst.New()
				for _, v := range lines[i]{
					t.Insert(v)
				}
				res[i] = t
			}
		}()
	}
	go func() {
		wg.Wait()
		close(jobs)
	}()

	return res
}