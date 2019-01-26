package passportChecker

import (
	"fmt"
)

type MapChecker struct {
	m1 map[string]struct{}
	m2 map[string]struct{}
}

//7.2gb(6 with each 1000000 gc), low cpu, fast resolves, handmade, easy to persist on disk, GC
func MakeMapChecker() *MapChecker {
	return &MapChecker{make(map[string]struct{}), make(map[string]struct{})}
}

func (c *MapChecker) Add(values []interface{}) error {
	for _, val := range values {
		key := fmt.Sprint(val)
		b := *c.bucket(key)
		b[key] = struct{}{}
	}
	return nil
}

func (c *MapChecker) Check(values []interface{}) ([]bool, error) {
	if len(values) == 0 {
		return make([]bool, 0), nil
	}
	result := make([]bool, 0, len(values))
	for _, val := range values {
		key := fmt.Sprint(val)
		b := *c.bucket(key)
		_, ok := b[key]
		result = append(result, ok)
	}
	return result, nil
}

func (c *MapChecker) bucket(k string) *map[string]struct{} {
	if []rune(k)[3] > '5' {
		return &c.m1
	}
	return &c.m2
}
