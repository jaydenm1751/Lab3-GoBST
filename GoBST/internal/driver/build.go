package driver

import (
	//stuff
	"gobst/internal/bst"
	"sync"
	"fmt"
)

func BuildTreesSequential(lines [][]int) []*bst.Tree{
	n := len(lines)
	trees := make([]*bst.Tree, n)
	for i, vals := range lines {
		t := bst.New()
		for _, v := range vals{
			t.Insert(v)
		}
		trees[i] = t
	}
	return trees
}

func BuildTreesParallel(lines [][]int, dataWorkers int) []*bst.Tree {
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
	for i := 0; i < n; i++ { jobs <- i }
	close(jobs)
	wg.Wait()

	for i, t := range res {
		if t == nil {
			panic(fmt.Errorf("build produced nil tree at index %d", i))
		}
	}


	return res
}