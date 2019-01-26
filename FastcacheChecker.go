package passportChecker

import (
	"fmt"
	"github.com/VictoriaMetrics/fastcache"
)

type FastCacheChecker struct {
	db *fastcache.Cache
}

// not able to store empty values
//10gb, mid cpu, 20µs resolves, easy to persist on disk, no GC
func MakeFastCacheChecker(db *fastcache.Cache) *FastCacheChecker {
	return &FastCacheChecker{db}
}

func (c *FastCacheChecker) Add(values []interface{}) error {
	for _, val := range values {
		c.db.Set([]byte(fmt.Sprint(val)), []byte("0"))
	}
	return nil
}

func (c *FastCacheChecker) Check(values []interface{}) ([]bool, error) {
	if len(values) == 0 {
		return make([]bool, 0), nil
	}
	result := make([]bool, 0, len(values))
	for _, val := range values {
		var b []byte
		r := c.db.Get(b, []byte(fmt.Sprint(val)))
		if r == nil {
			result = append(result, false)
		} else {
			result = append(result, true)
		}

	}
	return result, nil
}
