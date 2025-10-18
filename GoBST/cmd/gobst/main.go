package main

import (
	"flag"
	"fmt"
	"gobst/internal/bst"
	"gobst/internal/driver"
	"os"
	"sort"
)

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}

func openMust(path string) *os.File {
	f, err := os.Open(path)
	if err != nil {
		fatal(err)
	}
	return f
}

func printGroupsAndHashes(groups [][]int, hashes []int) {
	fmt.Println("== Results ==")
	for i := range groups {
		sort.Ints(groups[i])
		fmt.Printf("%d: %v\n", i, groups[i])
	}
	for i := range hashes {
		fmt.Printf("%d: %d\n", i, hashes[i])
	}
}

func main() {
	// Flags per spec
	hashWorkers := flag.Int("hash-workers", 1, "number of hash workers")
	dataWorkers := flag.Int("data-workers", 1, "number of data workers (tree build)")
	compWorkers := flag.Int("comp-workers", 1, "number of comparison workers")
	input := flag.String("input", "testdata/simple.txt", "path to input file")
	flag.Parse()

	// 1) Read & parse input
	f := openMust(*input)
	defer f.Close()

	lines, err := driver.ParseInput(f)
	if err != nil {
		fatal(err)
	}
	n := len(lines)
	if n == 0 {
		fmt.Println("== Results ==")
		return
	}

	// 2) Build trees (you said you’re keeping parallel build — that’s fine)
	var trees []*bst.Tree
	if *dataWorkers > 1 {
		trees = driver.BuildTreesParallel(lines, *dataWorkers) // MUST be exported
	} else {
		trees = driver.BuildTreesSequential(lines) // MUST be exported
	}

	// 3) Step 2 selection (per your mapping in the spec):
	//    -hash-workers=1 -data-workers=1        => sequential hashing (fallback to channel impl with 1 worker ok)
	//    -hash-workers=i -data-workers=1 (i>1)  => channel collector (Step 2A)
	//    -hash-workers=i -data-workers=i (i>1)  => global mutex map (Step 2B)
	var (
		buckets map[int][]int
		hashes  []int
	)

	switch {
	case *hashWorkers == 1 && *dataWorkers == 1:
		// simple sequential hashing: loop in main
		buckets = make(map[int][]int)
		hashes = make([]int, n)
		for id, t := range trees {
			h := t.HashValue()
			hashes[id] = h
			buckets[h] = append(buckets[h], id)
		}

	case *hashWorkers > 1 && *dataWorkers == 1:
		// Step 2A: per spec — hashing goroutines send (id,hash) to a single collector via channel
		buckets, hashes = driver.Step2Chan(trees, *hashWorkers)

	case *hashWorkers > 1 && *dataWorkers == *hashWorkers:
		// Step 2B: per spec — hash workers update shared map guarded by a single global mutex
		fmt.Printf("entering step2 mutexes\n")
		buckets, hashes = driver.Step2Mutexes(trees, *hashWorkers)

	default:
		// Sensible default: channel collector
		buckets, hashes = driver.Step2Chan(trees, *hashWorkers)
	}

	// 4) Step 3 selection:
	//fmt.Printf("exited switch\n")
	var groups [][]int

	if *compWorkers <= 1 {
		groups = driver.CompareSequential(lines, trees, buckets)
	} else {
		adj := driver.Step3Workers(trees, buckets, *compWorkers)
		groups = driver.AdjToGroups(adj)
	}

	// 5) Convert adj to groups (for the same output format you’ve been using) and print hashes
	printGroupsAndHashes(groups, hashes)

	fmt.Printf("Processed %d trees.\n", n)
}
