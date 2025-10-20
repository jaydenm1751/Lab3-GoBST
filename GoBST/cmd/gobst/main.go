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

// func printOutput(buckets map[int][]int,
// 	groups [][]int,
// 	//hashTime, hashGroupTime, compareTime time.Duration,
// ) {

// 	//hash time
// 	// fmt.Printf("hashTime: %.9f\n", hashTime.Seconds())
// 	// fmt.Printf("hashGroupTime: %.9f\n", hashGroupTime.Seconds())

// 	keys := make([]int, 0, len(buckets))
// 	for k := range buckets {
// 		if (len(buckets[k]) <= 1) {
// 			continue
// 		}
// 		keys = append(keys, k)
// 	}
// 	sort.Ints(keys)
// 	for _, h := range keys {
// 		ids := append([]int(nil), buckets[h]...)
// 		sort.Ints(ids)
// 		fmt.Printf("%d: ", h)
// 		for _, v := range ids {
// 			fmt.Printf("%d ", v)
// 		}
// 		fmt.Println()
// 	}

// 	//compare tree time
// 	// fmt.Printf("hashTime: %.9f\n", hashTime.Seconds())
// 	// fmt.Printf("compareTreeTime: %.9f\n", hashGroupTime.Seconds())

// 	for i := range groups {
// 		sort.Ints(groups[i])
// 		fmt.Printf("%d: %v\n", i, groups[i])
// 	}
// }

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
	var (
		buckets map[int][]int
		//hashes  []int
	)

	switch {
	case *hashWorkers == 1 && *dataWorkers == 1:
		// simple sequential hashing: loop in main
		buckets = make(map[int][]int)
		//hashes = make([]int, n)
		for id, t := range trees {
			h := t.HashValue()
			// hashes[id] = h
			buckets[h] = append(buckets[h], id)
		}

	case *hashWorkers > 1 && *dataWorkers == 1:
		// Step 2A: per spec — hashing goroutines send (id,hash) to a single collector via channel
		buckets = driver.Step2Chan(trees, *hashWorkers)

	case *hashWorkers > 1 && *dataWorkers == *hashWorkers:
		// Step 2B: per spec — hash workers update shared map guarded by a single global mutex
		// fmt.Printf("entering step2 mutexes\n")
		buckets = driver.Step2Mutexes(trees, *hashWorkers)

	default:
		// Sensible default: channel collector
		buckets = driver.Step2Chan(trees, *hashWorkers)
	}

	keys := make([]int, 0, len(buckets))
	for k := range buckets {
		if (len(buckets[k]) <= 1) {
			continue
		}
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, h := range keys {
		ids := append([]int(nil), buckets[h]...)
		sort.Ints(ids)
		fmt.Printf("%d: ", h)
		for _, v := range ids {
			fmt.Printf("%d ", v)
		}
		fmt.Println()
	}

	// 4) Step 3 selection:
	//fmt.Printf("exited switch\n")
	//groupIdx := 0
	var groups [][]int

	if *compWorkers <= 1 {
		for _, h := range keys {
			ids := append([]int(nil), buckets[h]...)
			sort.Ints(ids)
			gs := driver.CompareBucketSequential(trees, ids)
			groups = append(groups, gs...) // append classes
		}
	} else {
		adj := driver.Step3Simple(trees, buckets) // Step 3 implementation A
		// adj := driver.Step3Workers(trees, buckets, *compWorkers) //Step 3 implementation B
		groups = driver.AdjToGroups(adj)
	}
	
	//fmt.Printf("compareTreeTime: %.9f\n", compareTime.Seconds())
	groupIdx := 0
	if *compWorkers <= 1 {
		for _, g := range groups {
			if len(g) <= 1 { continue }
			fmt.Printf("group %d: ", groupIdx)
			for _, id := range g { 
				// if len(g) <= 1 { continue }
				fmt.Printf("%d ", id) 
			}
			fmt.Println()
			groupIdx++
		}
	} else {
		for _, h := range keys {
			if len(buckets[h]) <= 1 { continue }
			ids := append([]int(nil), buckets[h]...)
			sort.Ints(ids)

			inBucket := make(map[int]bool, len(ids))
			for _, id := range ids { inBucket[id] = true }

			printed := make(map[int]bool, len(ids)) // per-bucket
			for _, id := range ids {
				if printed[id] { continue }

				cls := make([]int, 0, len(groups[id]))
				for _, x := range groups[id] {
					if inBucket[x] { cls = append(cls, x) }
				}
				sort.Ints(cls)
				// mark if already printed
				for _, x := range cls { printed[x] = true }
				if len(cls) <= 1 { continue }

				fmt.Printf("group %d: ", groupIdx)
				for _, x := range cls { fmt.Printf("%d ", x) }
				fmt.Println()
				groupIdx++
			}
		}
	}

	// 5) Convert adj to groups (for the same output format you’ve been using) and print hashes
	//printOutput(groups, buckets)

	fmt.Printf("Processed %d trees.\n", n)
}
