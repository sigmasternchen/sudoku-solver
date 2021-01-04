package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sync/atomic"
	"time"
)

/*const size = 9
const blocks = 3
const blockSize = size / blocks*/

const size = 16
const blocks = 4
const blockSize = size / blocks

const maskZero = (1 << size) - 1

const guessChanSize = 1 << 20
const eliminatedChanSize = 1 << 15

type field [size][size][]int

var guesses = make(chan field, guessChanSize)
var solution = make(chan field, 1)

var eliminatedChan = make(chan int, eliminatedChanSize)

// 1 because of the first guess = starting point
var possibilities int64 = 1
var globalEliminated int64 = 0

func eliminate(field *field) *int {
	eliminated := 0

	// columns
	for x := 0; x < size; x++ {
		bitfield := maskZero

		for y := 0; y < size; y++ {
			if len(field[x][y]) == 1 {
				mask := 1 << (field[x][y][0] - 1)

				if bitfield&mask == 0 {
					// duplicate number
					return nil
				}

				bitfield &= mask ^ maskZero
			}
		}

		for y := 0; y < size; y++ {
			if len(field[x][y]) != 1 {
				length := len(field[x][y])
				for i := 0; i < length; i++ {
					if (bitfield & (1 << (field[x][y][i] - 1))) == 0 {
						eliminated++
						field[x][y][i] = field[x][y][length-1]
						field[x][y] = field[x][y][:length-1]
						length--
						i--
					}
				}

				if length == 0 {
					return nil
				}
			}
		}
	}

	// rows
	for y := 0; y < size; y++ {
		bitfield := maskZero

		for x := 0; x < size; x++ {
			if len(field[x][y]) == 1 {
				mask := 1 << (field[x][y][0] - 1)

				if bitfield&mask == 0 {
					// duplicate number
					return nil
				}

				bitfield &= mask ^ maskZero
			}
		}

		for x := 0; x < size; x++ {
			if len(field[x][y]) != 1 {
				length := len(field[x][y])
				for i := 0; i < length; i++ {
					if (bitfield & (1 << (field[x][y][i] - 1))) == 0 {
						eliminated++
						field[x][y][i] = field[x][y][length-1]
						field[x][y] = field[x][y][:length-1]
						length--
						i--
					}
				}

				if length == 0 {
					return nil
				}
			}
		}
	}

	// blocks
	for sx := 0; sx < blocks; sx++ {
		for sy := 0; sy < blocks; sy++ {
			bitfield := maskZero

			for x := sx * blockSize; x < (sx+1)*blockSize; x++ {
				for y := sy * blockSize; y < (sy+1)*blockSize; y++ {
					if len(field[x][y]) == 1 {
						mask := 1 << (field[x][y][0] - 1)

						if bitfield&mask == 0 {
							// duplicate number
							return nil
						}

						bitfield &= mask ^ maskZero
					}
				}
			}

			for x := sx * blockSize; x < (sx+1)*blockSize; x++ {
				for y := sy * blockSize; y < (sy+1)*blockSize; y++ {
					if len(field[x][y]) != 1 {
						length := len(field[x][y])
						for i := 0; i < length; i++ {
							if (bitfield & (1 << (field[x][y][i] - 1))) == 0 {
								eliminated++
								field[x][y][i] = field[x][y][length-1]
								field[x][y] = field[x][y][:length-1]
								length--
								i--
							}
						}

						if length == 0 {
							return nil
						}
					}
				}
			}
		}
	}

	return &eliminated
}

type result int

const (
	unsolved result = iota
	solved
	wrong
)

func check(field *field) result {
	for _, column := range field {
		for _, row := range column {
			switch len(row) {
			case 0:
				return wrong
			case 1:
				// check next
			default:
				return unsolved
			}
		}
	}
	return solved
}

func copyField(fieldToCopy field) field {
	var fieldCopy field
	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			fieldCopy[x][y] = make([]int, len(fieldToCopy[x][y]))
			copy(fieldCopy[x][y], fieldToCopy[x][y])
		}
	}

	return fieldCopy
}

func addGuesses(field field) {
	hasGuess := false
	minX := 0
	minY := 0
	minLength := size + 1

	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			length := len(field[x][y])
			if length > 1 && length < minLength {
				hasGuess = true
				minX = x
				minY = y
				minLength = length
			}
		}
	}

	if hasGuess {
		for _, value := range field[minX][minY] {
			newField := copyField(field)
			newField[minX][minY] = []int{value}

			atomic.AddInt64(&possibilities, 1)
			guesses <- newField
		}
	}
}

func removePossibility() {
	if atomic.AddInt64(&possibilities, -1) <= 0 {
		fmt.Println("no solution found")
		os.Exit(1)
	}
}

func worker() {
workerLoop:

	for {
		field := <-guesses
	eliminateLoop:
		for {
			eliminated := eliminate(&field)
			if eliminated == nil {
				removePossibility()
				continue workerLoop
			} else if *eliminated == 0 {
				break eliminateLoop
			} else {
				eliminatedChan <- *eliminated
			}
		}

		switch check(&field) {
		case wrong:
			removePossibility()
			continue workerLoop
		case solved:
			solution <- field
			// we don't need the worker anymore
			break workerLoop
		case unsolved:
			addGuesses(field)
			removePossibility()
			continue workerLoop
		}
	}
}

func all() []int {
	var list []int
	for i := 0; i < size; i++ {
		list = append(list, i+1)
	}
	return list
}

func readField(input string) field {
	var field field
	x := 0
	y := 0

	for _, c := range input {
		inc := false
		if c == ' ' || c == '-' {
			field[x][y] = all()
			inc = true
		} else if c >= '0' && c <= '9' {
			field[x][y] = []int{int(c - '0')}
			inc = true
		} else if c >= 'a' && c <= 'z' {
			field[x][y] = []int{int(c - 'a' + 10)}
			inc = true
		} else {
			// ignore
		}

		if inc {
			x++
			if x >= size {
				x = 0
				y++
			}
			if y >= size {
				break
			}
		}
	}

	for x = 0; x < size; x++ {
		for y = 0; y < size; y++ {
			if field[x][y] == nil {
				log.Fatal("not all cells filled")
			} else {
				for _, value := range field[x][y] {
					if value < 1 || value > size {
						log.Fatal("cell range exceeded")
					}
				}
			}
		}
	}

	return field
}

func printField(field field) {
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if len(field[x][y]) == 0 {
				fmt.Print("#")
			} else if len(field[x][y]) == 1 {
				value := field[x][y][0]
				if value < 10 {
					fmt.Printf("%c", rune(value+'0'))
				} else {
					fmt.Printf("%c", rune(value-10+'a'))
				}
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Print("\n")
	}
}

func eliminiatedCounter() {
	for {
		eliminiated := <-eliminatedChan
		globalEliminated += int64(eliminiated)
	}
}

/*const wikipedia = `
53--7----
6--195---
-98----6-
8---6---3
4--8-3--1
7---2---6
-6----28-
---419--5
----8--79
`

const medium = `
47-53----
-9-----8-
-1--2--5-
1--7--5-4
--39--1--
---65-9-3
95--172-6
28------1
7---64-9-
`

const hard = `
---29--4-
--31-5-26
--96-----
2----83--
1----98-5
-57------
768-----4
----6-2-9
----4---3
`

const veryDifficult = `
-21-6-4--
---5---9-
4----2--1
84--5----
1-------2
----4--75
7--6----4
-3---9---
--8-3-61-
`

const hardest = `
8--------
--36-----
-7--9-2--
-5---7---
----457--
---1---3-
--1----68
--85---1-
-9----4--
`*/

const test16x16 = `
c3-8---g7f---4-e
---ac-3-----g2-7
-b-d--8-5-9---f-
-9-4--2---gdb-53
--f-----27b---4-
4--------5-g9-7c
---e3----c--85--
-c951--e8-a4f---
---g48-56--a1ed-
--d9--b----c5---
5a-6f-e--------b
-e---7d2-----8--
a2-f85---g--4-e-
-4---a-3-2--6-1-
g-5c-----4-e2---
6-e---gfa---3-c5
`

func main() {
	field := readField(test16x16)
	printField(field)
	guesses <- field

	fmt.Println()

	threads := runtime.GOMAXPROCS(0)
	fmt.Printf("number of threads: %d\n", threads)
	for x := 0; x < threads; x++ {
		go worker()
	}

	go eliminiatedCounter()

	for {
		fmt.Printf("queued: %10d, remaining: %10d, eliminated: %d (+ %d)   \r", len(guesses), possibilities, globalEliminated, len(eliminatedChan))
		time.Sleep(10 * time.Millisecond)

		if len(solution) > 0 {
			break
		}
	}

	solution := <-solution

	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println("solved")

	printField(solution)
}
