package main

import (
	"errors"
	"fmt"
)

var (
	ErrNoSuchElement   = errors.New("нет такого элемента")
	ErrIndexOutOfBound = errors.New("индекс выходит за пределы допустимого диапазона")
)

type LinkedList struct {
	First *Node
	Last  *Node
	Size  int
}

type Node struct {
	Item int
	Next *Node
	Prev *Node
}

func NewLinkedList() *LinkedList {
	return &LinkedList{}
}

func (ll *LinkedList) Add(index int, item int) error {
	if index < 0 || index > ll.Size {
		return ErrIndexOutOfBound
	}

	if index == 0 {
		ll.AddFirst(item)
		return nil
	}

	if index == ll.Size {
		ll.AddLast(item)
		return nil
	}

	nextNode, _ := ll.Get(index)
	prevNode := nextNode.Prev

	newNode := &Node{Item: item}

	newNode.Prev = prevNode
	newNode.Next = nextNode

	prevNode.Next = newNode
	nextNode.Prev = newNode

	ll.Size++

	return nil
}

func (ll *LinkedList) AddFirst(item int) {
	newNode := &Node{Item: item}

	if ll.IsEmpty() {
		ll.First = newNode
		ll.Last = newNode
	} else {
		newNode.Next = ll.First
		ll.First.Prev = newNode
		ll.First = newNode
	}

	ll.Size++
}

func (ll *LinkedList) AddLast(item int) {
	newNode := &Node{Item: item}

	if ll.IsEmpty() {
		ll.First = newNode
		ll.Last = newNode
	} else {
		newNode.Prev = ll.Last
		ll.Last.Next = newNode
		ll.Last = newNode
	}

	ll.Size++
}

func (ll *LinkedList) Remove(index int) (*Node, error) {
	if index < 0 || index >= ll.Size {
		return nil, ErrIndexOutOfBound
	}

	if index == 0 {
		return ll.RemoveFirst()
	}

	if index == ll.Size-1 {
		return ll.RemoveLast()
	}

	currentNode, _ := ll.Get(index)

	currentNode.Prev.Next = currentNode.Next
	currentNode.Next.Prev = currentNode.Prev

	currentNode.Prev = nil
	currentNode.Next = nil

	ll.Size--

	return currentNode, nil
}

func (ll *LinkedList) RemoveFirst() (*Node, error) {
	if ll.IsEmpty() {
		return nil, ErrNoSuchElement
	}

	removedNode := ll.First

	if ll.Size == 1 {
		ll.First = nil
		ll.Last = nil
	} else {
		ll.First = removedNode.Next
		ll.First.Prev = nil
		removedNode.Next = nil
	}

	ll.Size--

	return removedNode, nil
}

func (ll *LinkedList) RemoveLast() (*Node, error) {
	if ll.IsEmpty() {
		return nil, ErrNoSuchElement
	}

	removedNode := ll.Last

	if ll.Size == 1 {
		ll.First = nil
		ll.Last = nil
	} else {
		ll.Last = removedNode.Prev
		ll.Last.Next = nil
		removedNode.Prev = nil
	}

	ll.Size--

	return removedNode, nil
}

func (ll *LinkedList) Get(index int) (*Node, error) {
	if index < 0 || index >= ll.Size {
		return nil, ErrIndexOutOfBound
	}

	currentNode := ll.First

	for range index {
		currentNode = currentNode.Next
	}

	return currentNode, nil
}

func (ll *LinkedList) Clear() {
	currentNode := ll.First

	for currentNode != nil {
		nextNode := currentNode.Next
		currentNode.Next = nil
		currentNode.Prev = nil
		currentNode = nextNode
	}

	ll.First = nil
	ll.Last = nil
	ll.Size = 0
}

func (ll *LinkedList) IsEmpty() bool {
	return ll.Size == 0
}

func main() {
	linkedList := NewLinkedList()

	fmt.Println("IsEmpty = ", linkedList.IsEmpty())

	linkedList.AddFirst(0)
	linkedList.AddFirst(1)
	if err := linkedList.Add(1, 2); err != nil {
		fmt.Printf("Failed to add num: %v", err)
	}
	if err := linkedList.Add(2, 3); err != nil {
		fmt.Printf("Failed to add num: %v", err)
	}

	printList(linkedList)

	removedNode, err := linkedList.Remove(2)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("removed:", removedNode.Item)

	removedFirst, err := linkedList.RemoveFirst()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("removed first:", removedFirst.Item)

	removedLast, err := linkedList.RemoveLast()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("removed last:", removedLast.Item)

	printList(linkedList)

	getNode, err := linkedList.Get(0)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("get:", getNode.Item)
	fmt.Println("IsEmpty = ", linkedList.IsEmpty())
}

func printList(ll *LinkedList) {
	currentNode := ll.First

	fmt.Print("List: ")
	for currentNode != nil {
		fmt.Print(currentNode.Item, ",")
		currentNode = currentNode.Next
	}
	fmt.Println()
}
