package passportChecker

import (
	"errors"
	"fmt"
	"github.com/labstack/gommon/log"
	"unicode/utf8"
)

type PrefixTreeChecker struct {
	t *PrefixTree
}

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
	val      Value
	children []node
}

type Value uint8

var Symbols []rune
var RuneDict map[rune]Value

func init() {
	Symbols = []rune{
		' ', '-', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', 'Б', 'А', 'Е', 'Г', 'Ж', 'М', 'B', 'T', 'O', 'Ю', 'Щ', 'О', 'Д', 'С', 'V', 'Т', 'Н', 'Л', 'X', 'Ч', 'Я', 'A', 'П', 'Ф', 'К', 'C', 'M', 'Р', 'Х', 'N', 'Ш', 'В', 'И', 'I', 'З', 'K', 'У',
	}
	RuneDict = make(map[rune]Value)
	for i, r := range Symbols {
		RuneDict[r] = Value(i)
	}
}

func (t *PrefixTree) Add(s string) error {
	chain, err := stringToChain(s)
	if err != nil {
		return err
	}
	if len(chain) == 0 {
		return nil
	}
	n := t.root
	for _, v := range chain {
		n = n.findChildOrNew(v)
	}
	return nil
}

func (t *PrefixTree) Check(s string) (bool, error) {
	chain, err := stringToChain(s)
	if err != nil {
		return false, err
	}
	if len(chain) == 0 {
		return false, nil
	}
	n := t.root
	for _, v := range chain {
		n = n.findChild(v)
		if n == nil {
			return false, nil
		}
	}
	return true, nil
}

func stringToChain(s string) ([]Value, error) {
	chain := make([]Value, 0, utf8.RuneCountInString(s))
	for _, r := range s {
		if v, ok := RuneDict[r]; ok {
			chain = append(chain, v)
		} else {
			return nil, errors.New("unsupported symbol:" + string(r))
		}
	}
	return chain, nil
}

func (n *node) findChildOrNew(v Value) *node {
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

func (n *node) findChild(v Value) *node {
	for i, child := range n.children {
		if child.val == v {
			return &n.children[i]
		}
	}
	return nil
}
