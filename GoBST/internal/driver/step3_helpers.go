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


func CompareSequential(lines [][]int, trees []*bst.Tree, buckets map[int][]int) [][]int {
	n := len(lines)
	groups := make([][]int, n)
	visited := make([]bool, n)
	for _, list := range buckets {
		for _, i := range list {
			if visited[i] {
				continue;
			}
			visited[i] = true
			eq := []int{i}
			for _, j := range list {
				if i == j || visited[j] {
					continue
				}
				if trees[i].Equal(trees[j]){
					visited[j] = true
					eq = append(eq, j)
				}
			}
			//sort.Ints(eq)
			for _, k := range eq {
				// groups[k] = append([]int(nil), eq...)
				groups[k] = eq
			}
		}
	}
	// fmt.Println("== Step 1 Results ==")
	// for i := 0; i < n; i++ {
	// 	///sort.Ints(groups[i])
	// 	fmt.Printf("%d: %v\n", i, groups[i])
	// 	//fmt.Printf("%d: hash=%03d identical=%v\n", i, hashes[i], groups[i])
	// }
	// for i := 0; i < n; i++ {
	// 	///sort.Ints(groups[i])
	// 	fmt.Printf("%d: %v\n", i , hashes[i])
	// 	//fmt.Printf("%d: hash=%03d identical=%v\n", i, hashes[i], groups[i])
	// }

	return groups
}
