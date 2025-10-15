package main

import (
	"flag"
	"fmt"
	"gobst/internal/driver"
)

func main() {
	// Accept worker flags for future steps, unused for Step 1.
	hashWorkers := flag.Int("hash-workers", 1, "number of hash workers (ignored in step 1)")
	dataWorkers := flag.Int("data-workers", 1, "number of data workers (ignored in step 1)")
	compWorkers := flag.Int("comp-workers", 1, "number of compare workers (ignored in step 1)")
	input := flag.String("input", "testdata/simple.txt", "path to input file")

	flag.Parse()
	_ = hashWorkers
	_ = dataWorkers
	_ = compWorkers

	hashes, err := driver.Sequential(*input)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Processed %d trees (sequential baseline).\n", len(hashes))
}
