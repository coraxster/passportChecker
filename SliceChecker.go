package passportChecker

import (
	"fmt"
)

type SliceChecker struct {
	sl []interface{}
}

// extremely not optimal, just for fun :) (btw, 4.5gb mem)
func MakeSliceChecker(n int) *SliceChecker {
	return &SliceChecker{make([]interface{}, n)}
}

func (c *SliceChecker) Add(values []interface{}) error {
	c.sl = append(c.sl, values...)
	return nil
}

func (c *SliceChecker) Check(values []interface{}) ([]bool, error) {
	if len(values) == 0 {
		return make([]bool, 0), nil
	}
	result := make([]bool, len(values))
	for _, sv := range c.sl {
		svStr := fmt.Sprint(sv)
		for i, val := range values {
			if svStr == fmt.Sprint(val) {
				result[i] = true
				break
			}
		}
	}
	return result, nil
}
