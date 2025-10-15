package driver

import (
	"fmt"
	"gobst/internal/bst"
	"os"
)

// Step1Sequential builds trees, hashes, dedups, and prints identical-ID lists.
// Returns: slice of hash strings aligned with tree indexes.

func Sequential(inputPath string) ([]int, error) {
	f, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	lines, err := parseInput(f)
	if err != nil { return nil, err }

	n := len(lines)
	trees := make([]*bst.Tree, n)
	hashes := make([]int, n)
	for i, vals := range lines {
		t := bst.New()
		for _, v := range vals{
			t.Insert(v)
		}
		trees[i] = t
	}

	hashStore := make(map[int][]int)
	for i, t := range trees{
		h := t.HashValue()
		hashes[i] = h
		hashStore[h] = append(hashStore[h], i)
	}

	groups := make([][]int, n)
	visited := make([]bool, n)
	for _, list := range hashStore {
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
	fmt.Println("== Step 1 Results ==")
	for i := 0; i < n; i++ {
		///sort.Ints(groups[i])
		fmt.Printf("%d: %v\n", i, groups[i])
		//fmt.Printf("%d: hash=%03d identical=%v\n", i, hashes[i], groups[i])
	}
	for i := 0; i < n; i++ {
		///sort.Ints(groups[i])
		fmt.Printf("%d: %v\n", i , hashes[i])
		//fmt.Printf("%d: hash=%03d identical=%v\n", i, hashes[i], groups[i])
	}

	return hashes, nil

}