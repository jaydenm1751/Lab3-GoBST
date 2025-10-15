package driver

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

func parseInput(r io.Reader) ([][]int, error){
	var groups [][]int
	sc := bufio.NewScanner(r)
	for sc.Scan(){
		line := strings.TrimSpace(sc.Text())
		if line == ""{
			groups = append(groups, []int{})
			continue
		}
		fields := strings.Fields(line)
		nums := make([]int, 0, len(fields))
		for _, f := range fields {
			x, err := strconv.Atoi(f)
			if err != nil {
				return nil, err
			}
			nums = append(nums, x)
		}
		groups = append(groups, nums)
	}
	return groups, sc.Err()
}