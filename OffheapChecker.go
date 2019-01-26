package passportChecker

import (
	"fmt"
	"github.com/glycerine/offheap"
	"log"
)

//not working. >10gb for 10 000 000 keys
type OffheapChecker struct {
	h *offheap.StringHashTable
}

func MakeOffheapChecker(h *offheap.StringHashTable) *OffheapChecker {
	return &OffheapChecker{h}
}

func (c *OffheapChecker) Add(values []interface{}) error {
	for _, val := range values {
		if c.h.InsertStringKey(fmt.Sprint(val), 0) == false {
			log.Println("problem with inserting")
		}
	}
	return nil
}

func (c *OffheapChecker) Check(values []interface{}) ([]bool, error) {
	if len(values) == 0 {
		return make([]bool, 0), nil
	}
	result := make([]bool, 0, len(values))
	for _, val := range values {
		_, r := c.h.LookupStringKey(fmt.Sprint(val))
		if !r {
			result = append(result, false)
		} else {
			result = append(result, true)
		}
	}
	return result, nil
}
