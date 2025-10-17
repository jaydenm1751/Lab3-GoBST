package driver

import (
	"sync"
	"gobst/internal/bst"
)



func Step3Simple(trees []*bst.Tree, buckets map[int][]int) [][]bool {
	n := len(trees)
	adj := MakeAdj(n)

	t_map := TreesIndex(trees)
	
	var wg sync.WaitGroup
	var lock sync.Mutex

	for _, ids := range buckets {
		if len(ids) < 2 { continue }
		for _, p := range BuildPairs(ids){
			wg.Add(1)
			go func(p pair){
				defer wg.Done()
				if t_map[p.i].Equal(t_map[p.j]){
					lock.Lock()
					adj[p.i][p.j], adj[p.j][p.i] = true, true
					lock.Unlock()
				}
			}(p)
		}
	}
	wg.Wait()
	return adj
}