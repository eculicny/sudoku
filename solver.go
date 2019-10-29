package main

import (
	"bytes"
	"fmt"
	"log"
	"math"
)

type pair struct {
	row int
	col int
}

func printGrid(grid [][]int) {
	l := len(grid)
	rt := math.Sqrt(float64(l))

	for i, row := range grid {
		if i%int(rt) == 0 {
			for i := 0; i < l+int(rt); i++ {
				fmt.Print("--")
			}
			fmt.Println()
		}

		for j, val := range row {
			if j%int(rt) == 0 {
				fmt.Print("| ")
			}
			fmt.Printf("%v ", val)
		}
		fmt.Println("|")
	}

	for i := 0; i < l+int(rt); i++ {
		fmt.Print("--")
	}
}

func validate(grid [][]int, logger *log.Logger) (valid bool, complete bool) {
	l := len(grid)
	rt := math.Sqrt(float64(l))

	var zeroInd []pair
	rowCounts := make([][]int, l)
	colCounts := make([][]int, l)
	blockCounts := make([][]int, l)
	for i := 0; i < l; i++ {
		rowCounts[i] = make([]int, l+1)
		colCounts[i] = make([]int, l+1)
		blockCounts[i] = make([]int, l+1)
	}

	for i := 0; i < l; i++ {
		for j := 0; j < l; j++ {
			blockIndex := int(math.Floor(float64(j)/rt)*rt + math.Floor(float64(i)/rt))
			val := grid[i][j]

			if val != 0 {
				if rowCounts[i][val] >= 1 {
					return false, false
				} else if colCounts[j][val] >= 1 {
					return false, false
				} else if blockCounts[blockIndex][val] >= 1 {
					return false, false
				}
			} else {
				zeroInd = append(zeroInd, pair{i, j})
			}

			rowCounts[i][val]++
			colCounts[j][val]++
			blockCounts[blockIndex][val]++
		}
	}

	fmt.Println(len(zeroInd))
	fmt.Println(zeroInd)
	if len(zeroInd) != 0 {
		ind := zeroInd[0]
		for i := 1; i <= l; i++ {
			if rowCounts[ind.row][i] == 0 && colCounts[ind.col][i] == 0 {
				grid[ind.row][ind.col] = i
				valid, complete := validate(grid, logger)
				if complete {
					fmt.Println("Completed grid achieved. Returning success")
					return true, true
				} else if valid {
					fmt.Println("Grid is valid. Continuing to solve")
					return validate(grid, logger)
				} else {
					fmt.Println("Invalid grid generated. Resetting index")
					grid[ind.row][ind.col] = 0
				}
			} else {
				fmt.Printf("Can't place %v at %v,%v\n", i, ind.row, ind.col)
			}
		}

		fmt.Println("No valid grids found. Returning failure")
		return false, false
	}

	fmt.Println("Hit final return (should be success)")
	return true, true
}

func main() {

	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "logger: ", log.Lshortfile)
	)

	//logger.Print("Beginning...")

	grid := [][]int{
		{1, 0, 0, 0, 6, 0, 0, 0, 0},
		{7, 0, 0, 5, 0, 3, 0, 0, 0},
		{6, 9, 0, 0, 0, 0, 0, 3, 0},
		{5, 0, 0, 2, 0, 0, 0, 7, 0},
		{9, 0, 0, 1, 7, 4, 0, 0, 5},
		{0, 4, 0, 0, 0, 6, 0, 0, 3},
		{0, 1, 0, 0, 0, 0, 0, 6, 2},
		{0, 0, 0, 3, 0, 7, 0, 0, 4},
		{0, 0, 0, 0, 1, 0, 0, 0, 9}}

	fmt.Println(validate(grid, logger))
	printGrid(grid)
	//logger.Print("Completed run")
}

/*
0,0 1
0,4 6
1,0 7
1,3 5
1,5 3
2,0 6
2,1 9
2,7 3
3,0 5
3,3 2
3,7 7
4,0 9
4,3 1
4,4 7
4,5 4
4,8 5
5,1 4
5,5 6
5,8 3
6,1 1
6,7 6
6,8 2
7,3 3
7,5 7
7,8 4
8,4 1
8,8 9
*/
