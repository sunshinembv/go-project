package main

import (
	"errors"
	"fmt"
)

var (
	NoSuchElementErr = errors.New("нет такого элемента")
)

func binarySearch(input []int, num int) (int, error) {
	first := 0
	last := len(input) - 1

	for first <= last {
		mid := first + (last-first)/2

		if input[mid] == num {
			return mid, nil
		}

		if input[mid] > num {
			last = mid - 1
		} else {
			first = mid + 1
		}
	}

	return -1, NoSuchElementErr
}

func main() {
	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	for i := range len(input) {
		res, err := binarySearch(input, i+1)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(res)
	}
}
