package driver

import (
	"gobst/internal/bst"
	"sort"
)

type pair struct {
	i, j int
}

func makeAdj(n int) [][]bool {
	adj := make([][]bool, n)
	for i := range adj{
		row := make([]bool, n)
		row[i] = true // (i, i) == true
		adj[i] = row
	}
	return adj
}

func TreesIndex(trees []*bst.Tree) map[int]*bst.Tree {
	m := make(map[int]*bst.Tree, len(trees))
	for id, t := range trees { m[id] = t }
	return m
}

func buildPairs(ids []int) []pair {
	p := make([]pair, 0, len(ids) * (len(ids) - 1) / 2)
	for i := 0; i < len(ids); i++{
		for j := i + 1; j < len(ids); j++ {
			p = append(p, pair{ids[i], ids[j]})
		}
	}
	return p
}
