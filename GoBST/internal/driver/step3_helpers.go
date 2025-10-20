package driver

import (
	"gobst/internal/bst"
	"sort"
	// "fmt"
)

type pair struct {
	i, j int
}

func MakeAdj(n int) [][]bool {
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

func BuildPairs(ids []int) []pair {
	p := make([]pair, 0, len(ids) * (len(ids) - 1) / 2)
	for i := 0; i < len(ids); i++{
		for j := i + 1; j < len(ids); j++ {
			p = append(p, pair{ids[i], ids[j]})
		}
	}
	return p
}

func AdjToGroups(adj [][]bool) [][]int {
    n := len(adj)
    groups := make([][]int, n)
    for i := 0; i < n; i++ {
        for j := 0; j < n; j++ {
            if adj[i][j] {
                groups[i] = append(groups[i], j)
            }
        }
        sort.Ints(groups[i])
    }
    return groups
}


func CompareBucketSequential(trees []*bst.Tree, ids []int) [][]int {
	if len(ids) == 0 { 
		return nil 
	} else if len(ids) == 1 {
		return [][]int{{ids[0]}}
	}
	sort.Ints(ids)

	used := make(map[int]bool, len(ids))
	groups := make([][]int, 0)

	for _, i := range ids {
		if used[i] { continue }
		used[i] = true
		cls := []int{i}
		for _, j := range ids {
			if j == i || used[j] { continue }
			if trees[i].Equal(trees[j]) {
				used[j] = true
				cls = append(cls, j)
			}
		}
		sort.Ints(cls)
		groups = append(groups, cls)
	}
	return groups
}
