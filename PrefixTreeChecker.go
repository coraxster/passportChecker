package passportChecker

import (
	"fmt"
	"github.com/labstack/gommon/log"
)

type PrefixTreeChecker struct {
	t *PrefixTree
}

//8.5gb (7.1 with each 1000000 gc), mid cpu, fast resolves, handmade, hard to persist on disk, GC
// looks like doesnt fit here
func MakePrefixTreeChecker(t *PrefixTree) *PrefixTreeChecker {
	return &PrefixTreeChecker{t}
}

func (c *PrefixTreeChecker) Add(values []interface{}) error {
	for _, val := range values {
		if err := c.t.Add(fmt.Sprint(val)); err != nil {
			log.Print(err.Error())
		}
	}
	return nil
}

func (c *PrefixTreeChecker) Check(values []interface{}) ([]bool, error) {
	if len(values) == 0 {
		return make([]bool, 0), nil
	}
	result := make([]bool, 0, len(values))
	for _, val := range values {
		r, err := c.t.Check(fmt.Sprint(val))
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

type PrefixTree struct {
	root *node
}

func MakePrefixTree() *PrefixTree {
	return &PrefixTree{&node{0, make([]node, 0)}}
}

type node struct {
	val      rune
	children []node
}

func (t *PrefixTree) Add(s string) error {

	if len(s) == 0 {
		return nil
	}
	n := t.root
	for _, v := range s {
		n = n.findChildOrNew(v)
	}
	return nil
}

func (t *PrefixTree) Check(s string) (bool, error) {
	if len(s) == 0 {
		return false, nil
	}
	n := t.root
	for _, v := range s {
		n = n.findChild(v)
		if n == nil {
			return false, nil
		}
	}
	return true, nil
}

func (n *node) findChildOrNew(v rune) *node {
	if found := n.findChild(v); found != nil {
		return found
	}
	if n.children == nil {
		n.children = make([]node, 1)
		n.children[0] = node{v, nil}
	} else {
		n.children = append(n.children, node{v, nil})
	}
	return &n.children[len(n.children)-1]
}

func (n *node) findChild(v rune) *node {
	for i, child := range n.children {
		if child.val == v {
			return &n.children[i]
		}
	}
	return nil
}
