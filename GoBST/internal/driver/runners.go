package driver

import (
	"gobst/internal/bst"
)

type ParallelRes struct {
	Buckets map[int][]int
	Hashes []int
	Trees []*bst.Tree
}


func Step2Run(lines [][]int, dataWorkers, hashWorkers int) (ParallelRes) {
	var trees []*bst.Tree
	trees = buildParallel(lines, dataWorkers)
	
	var buckets map[int][]int
	var hashes []int

	if hashWorkers > 1 {
		buckets, hashes = Step2Mutexes(trees, hashWorkers)
	} else {
		hashWorkers = 1
		buckets, hashes = Step2Chan(trees)
	}
	return ParallelRes{Buckets: buckets, Hashes: hashes, Trees: trees}
}

func Step3Run(buckets map[int][]int, hashes []int, trees []*bst.Tree, compWorkers int) [][]int {
	n := len(trees)
	groups := make([][]int, n)
	return groups
}