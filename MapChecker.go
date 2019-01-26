package passportChecker

import (
	"errors"
	"fmt"
	"log"
)

type MapChecker struct {
	m1 map[[10]Value]struct{}
	m2 map[[10]Value]struct{}
}

func MakeMapChecker() *MapChecker {
	return &MapChecker{make(map[[10]Value]struct{}), make(map[[10]Value]struct{})}
}

func (c *MapChecker) Add(values []interface{}) error {
	for _, val := range values {
		vs, err := stringToChain(fmt.Sprint(val))
		if err != nil {
			log.Print(err.Error())
			continue
		}
		if len(vs) > 10 {
			log.Print("too big string:" + fmt.Sprint(val))
			continue
		}
		var key [10]Value
		for i, v := range vs {
			key[i] = v
		}
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
		vs, err := stringToChain(fmt.Sprint(val))
		if err != nil {
			return nil, err
		}
		if len(vs) > 12 {
			return nil, errors.New("too big string:" + fmt.Sprint(val))
		}
		var key [10]Value
		for i, v := range vs {
			key[i] = v
		}
		b := *c.bucket(key)
		_, ok := b[key]
		result = append(result, ok)
	}
	return result, nil
}

func (c *MapChecker) bucket(k [10]Value) *map[[10]Value]struct{} {
	if k[4] > Value(25) {
		return &c.m1
	}
	return &c.m2
}
