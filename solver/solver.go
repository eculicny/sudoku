package solver

import (
	"fmt"
	"log"
	"math"
	"sync"
	// TODO use context instead of channelsfor threading
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

	fmt.Println()
}

func BruteForce(grid [][]int, logger *log.Logger) (valid bool, complete bool) {
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

	//fmt.Println(len(zeroInd))
	//fmt.Println(zeroInd)
	if len(zeroInd) != 0 {
		ind := zeroInd[0]
		for i := 1; i <= l; i++ {
			if rowCounts[ind.row][i] == 0 && colCounts[ind.col][i] == 0 {
				grid[ind.row][ind.col] = i
				valid, complete := BruteForce(grid, logger)
				if complete {
					//fmt.Println("Completed grid achieved. Returning success")
					return true, true
				} else if valid {
					//fmt.Println("Grid is valid. Continuing to solve")
					return BruteForce(grid, logger)
				} else {
					//fmt.Println("Invalid grid generated. Resetting index")
					grid[ind.row][ind.col] = 0
				}
			} else {
				//fmt.Printf("Can't place %v at %v,%v\n", i, ind.row, ind.col)
			}
		}

		//fmt.Println("No valid grids found. Returning failure")
		return false, false
	}

	//fmt.Println("Hit final return (should be success)")
	return true, true
}

func bruteForceThread(grid [][]int, idx pair, halt chan bool, out chan [][]int, logger *log.Logger, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 1; i <= len(grid[0]); i++ {
		select {
		case _, ok := <-halt:
			if ok {
				//fmt.Printf("Halting routine %v\n", idx)
				return
			} else { // probably not the best way to stop a go routine
				//fmt.Printf("Halting channel already closed for routine %v\n", idx)
				return
			}
		default:
			{
				grid[idx.row][idx.col] = i
				_, complete := BruteForce(grid, logger)
				if complete {
					//fmt.Printf("Solution found in routine %v\n", idx)
					out <- grid
					//fmt.Printf("Wrote solution to output channel %v\n", idx)
					return
				}
			}
		}
	}

	return
}

func BruteForceParallel(grid [][]int, tasks int, logger *log.Logger) (valid bool, complete bool) {
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
	halt := make(chan bool, 10)
	output := make(chan [][]int, 20)
	var solverwg sync.WaitGroup
	//var overseerwg sync.WaitGroup

	// TODO handle case where more threads than zeroes
	d := int(math.Floor(float64(len(zeroInd)) / float64(tasks)))

	fmt.Printf("len: %v, d: %v\n", len(zeroInd), d)
	for i := 0; i < len(zeroInd); i = i + d {
		fmt.Printf("Starting thread %v\n", i)
		// TODO move to method make a copy of the grid for the goroutine
		gridThread := make([][]int, len(grid))
		for i := range grid {
			gridThread[i] = make([]int, len(grid[i]))
			copy(gridThread[i], grid[i])
		}

		solverwg.Add(1)
		go bruteForceThread(gridThread, zeroInd[i], halt, output, logger, &solverwg)
	}

	// TODO figure out how to handle the case when no solution exists

	// wait until a solution is found
	soln := <-output

	// probably redundant w/ channel closure
	halt <- true
	close(halt)

	solverwg.Wait()
	close(output)

	for i := range grid {
		grid[i] = make([]int, len(soln[i]))
		copy(grid[i], soln[i])
	}
	return true, true
}

// TODO move to test
// func main() {

// 	var (
// 		buf    bytes.Buffer
// 		logger = log.New(&buf, "logger: ", log.Lshortfile)
// 	)

// 	logger.Print("Beginning...")

// 	grid := [][]int{
// 		{1, 0, 0, 0, 6, 0, 0, 0, 0},
// 		{7, 0, 0, 5, 0, 3, 0, 0, 0},
// 		{6, 9, 0, 0, 0, 0, 0, 3, 0},
// 		{5, 0, 0, 2, 0, 0, 0, 7, 0},
// 		{9, 0, 0, 1, 7, 4, 0, 0, 5},
// 		{0, 4, 0, 0, 0, 6, 0, 0, 3},
// 		{0, 1, 0, 0, 0, 0, 0, 6, 2},
// 		{0, 0, 0, 3, 0, 7, 0, 0, 4},
// 		{0, 0, 0, 0, 1, 0, 0, 0, 9}}

// 	gridBF := make([][]int, len(grid))
// 	for i := range grid {
// 		gridBF[i] = make([]int, len(grid[i]))
// 		copy(gridBF[i], grid[i])
// 	}

// 	gridBFP := make([][]int, len(grid))
// 	for i := range grid {
// 		gridBFP[i] = make([]int, len(grid[i]))
// 		copy(gridBFP[i], grid[i])
// 	}

// 	var start time.Time

// 	start = time.Now()
// 	fmt.Println(BruteForce(gridBF, logger))
// 	t1 := time.Now().Sub(start)

// 	start = time.Now()
// 	fmt.Println(BruteForceParallel(gridBFP, 10, logger))
// 	t2 := time.Now().Sub(start)

// 	printGrid(gridBF)
// 	printGrid(gridBFP)

// 	fmt.Printf("Brute force runtime: %v\n", t1)
// 	fmt.Printf("Brute force runtime: %v\n", t2)
// }
