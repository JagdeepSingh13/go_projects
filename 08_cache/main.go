package main

import "fmt"

const size = 5

type Node struct {
	Val   string
	Left  *Node
	Right *Node
}

// Head and Tail are empty nodes which point at
// first and last values resp.
type Queue struct {
	Head   *Node
	Tail   *Node
	Length int
}

type Hash map[string]*Node

type Cache struct {
	Queue Queue
	Hash  Hash
}

func NewCache() Cache {
	return Cache{Queue: NewQueue(), Hash: Hash{}}
}

func NewQueue() Queue {
	head := &Node{}
	tail := &Node{}

	// it means an empty queue as there is noting to right of head
	head.Right = tail
	tail.Left = head

	return Queue{Head: head, Tail: tail}
}

func (c *Cache) Check(str string) {
	node := &Node{}

	// check if value already exists in the Hash
	if val, ok := c.Hash[str]; ok {
		node = c.Remove(val)
	} else {
		node = &Node{Val: str}
	}

	c.Add(node)
	c.Hash[str] = node
}

func (c *Cache) Remove(n *Node) *Node {
	fmt.Printf("remove: %s\n", n.Val)

	left := n.Left
	right := n.Right

	left.Right = right
	right.Left = left

	c.Queue.Length -= 1
	delete(c.Hash, n.Val)

	return n
}

func (c *Cache) Add(n *Node) {
	fmt.Printf("add: %s\n", n.Val)

	tmp := c.Queue.Head.Right

	c.Queue.Head.Right = n
	n.Left = c.Queue.Head
	n.Right = tmp
	tmp.Left = n

	c.Queue.Length++

	if c.Queue.Length > size {
		c.Remove(c.Queue.Tail.Left)
	}
}

func (c *Cache) Display() {
	c.Queue.Display()
}

func (q *Queue) Display() {
	node := q.Head.Right

	fmt.Printf("%d - [", q.Length)

	for i := 0; i < q.Length; i++ {
		fmt.Printf("{%s}", node.Val)

		if i < q.Length-1 {
			fmt.Printf("<-->")
		}
		node = node.Right
	}
	fmt.Println("]")
}

func main() {
	fmt.Println("START CACHE")

	cache := NewCache()
	for _, word := range []string{"parrot", "avocado", "tree", "parrot", "pet", "avocado"} {
		cache.Check(word)
		cache.Display()
	}
}
