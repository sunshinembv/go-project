package main

import (
	"errors"
	"fmt"
)

var (
	ErrStackEmpty = errors.New("stack is empty")
)

type Stack struct {
	Values []int
}

func NewStack() *Stack {
	return &Stack{
		make([]int, 0),
	}
}

func (s *Stack) Push(value int) {
	s.Values = append(s.Values, value)
}

func (s *Stack) Pop() (int, error) {
	if len(s.Values) == 0 {
		return 0, ErrStackEmpty
	}

	lastIndex := len(s.Values) - 1
	value := s.Values[lastIndex]

	s.Values = s.Values[:lastIndex]
	return value, nil
}

func (s *Stack) Peek() (int, error) {
	if len(s.Values) == 0 {
		return 0, ErrStackEmpty
	}
	return s.Values[len(s.Values)-1], nil
}

func main() {
	stack := NewStack()

	stack.Push(1)
	stack.Push(3)
	stack.Push(2)

	fmt.Println("Stack:", stack.Values)
	fmt.Println()

	value, err := stack.Peek()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Peek:", value)
	fmt.Println("Stack:", stack.Values)
	fmt.Println()

	value, err = stack.Pop()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Pop:", value)
	fmt.Println("Stack:", stack.Values)
	fmt.Println()

	value, err = stack.Pop()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Pop:", value)
	fmt.Println("Stack:", stack.Values)
}
