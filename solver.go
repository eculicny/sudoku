package main

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"sync"
	"time"
)

// TODO stop using multidimensional array and use single array
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

func bruteForce(grid [][]int, logger *log.Logger) (valid bool, complete bool) {
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
				valid, complete := bruteForce(grid, logger)
				if complete {
					fmt.Println("Completed grid achieved. Returning success")
					return true, true
				} else if valid {
					fmt.Println("Grid is valid. Continuing to solve")
					return bruteForce(grid, logger)
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

func bruteForceThread(grid [][]int, idx pair, halt chan bool, out chan [][]int, logger *log.Logger, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 1; i <= len(grid[0]); i++ {
		select {
		case _, ok := <-halt:
			if ok {
				fmt.Printf("Halting routine %v\n", idx)
				return
			} else { // probably not the best way to stop a go routine
				fmt.Printf("Halting channel already closed for routine %v\n", idx)
				return
			}
		default:
			{
				grid[idx.row][idx.col] = i
				_, complete := bruteForce(grid, logger)
				if complete {
					out <- grid
					return
				}
			}
		}
	}

	return
}

func bruteForceParallel(grid [][]int, tasks int, logger *log.Logger) (valid bool, complete bool) {
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

	// this is not going to work as originally intended since channels are like queues
	halt := make(chan bool)
	output := make(chan [][]int)
	var solverwg sync.WaitGroup
	//var overseerwg sync.WaitGroup

	// TODO handle case where more threads than zeroes
	d := int(math.Floor(float64(len(zeroInd)) / float64(tasks)))

	for i := 0; i < len(zeroInd); i = i + d {
		// TODO move to method make a copy of the grid for the goroutine
		gridThread := make([][]int, len(grid))
		for i := range grid {
			gridThread[i] = make([]int, len(grid[i]))
			copy(gridThread[i], grid[i])
		}

		solverwg.Add(1)
		go bruteForceThread(gridThread, zeroInd[i], halt, output, logger, &solverwg)
	}

	// central routine to close output channel
	// overseerwg.Add(1)
	// go func(output chan [][]int, logger *log.Logger, wg *sync.WaitGroup, mywg *sync.WaitGroup) {
	// 	fmt.Println("Waiting")
	// 	wg.Wait()
	// 	fmt.Println("Waitgroup finished. Closing output channel.")
	// 	close(output)
	// 	mywg.Done()
	// }(output, logger, &solverwg, &overseerwg)

	// wait until a solution is found
	for soln := range output {
		halt <- true // probably redundant w/ channel closure
		close(halt)
		//overseerwg.Wait()

		solverwg.Wait()
		close(output) // this seems like bad practice :)

		grid = soln
		return true, true
	}

	return false, false
}

func main() {

	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "logger: ", log.Lshortfile)
	)

	logger.Print("Beginning...")

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

	gridBF := make([][]int, len(grid))
	for i := range grid {
		gridBF[i] = make([]int, len(grid[i]))
		copy(gridBF[i], grid[i])
	}

	gridBFP := make([][]int, len(grid))
	for i := range grid {
		gridBFP[i] = make([]int, len(grid[i]))
		copy(gridBFP[i], grid[i])
	}

	var start time.Time
	var t time.Duration

	// start = time.Now()
	// fmt.Println(bruteForce(gridBF, logger))
	// t = time.Now().Sub(start)
	// printGrid(gridBF)
	// fmt.Printf("Brute force runtime: %v\n", t)

	start = time.Now()
	fmt.Println(bruteForceParallel(gridBFP, 5, logger))
	t = time.Now().Sub(start)
	printGrid(gridBFP)
	fmt.Printf("Brute force runtime: %v\n", t)
}
