package passportChecker

import (
	"github.com/seiflotfy/cuckoofilter"
)

type CuckooChecker struct {
	cf *cuckoo.Filter
}

func MakeCuckooChecker(cf *cuckoo.Filter) (*CuckooChecker, error) {
	return &CuckooChecker{cf}, nil
}

func (c *CuckooChecker) Add(values []string) error {
	for _, val := range values {
		c.cf.InsertUnique([]byte(val))
	}
	return nil
}

func (s *CuckooChecker) Check(values []string) ([]bool, error) {
	if len(values) == 0 {
		return make([]bool, 0), nil
	}
	result := make([]bool, 0, len(values))
	for _, val := range values {
		result = append(result, s.cf.Lookup([]byte(val)))
	}
	return result, nil
}
