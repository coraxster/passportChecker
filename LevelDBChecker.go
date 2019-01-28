package passportChecker

import (
	"fmt"
	"github.com/labstack/gommon/log"
	"github.com/syndtr/goleveldb/leveldb"
)

// медленно пишет, 100% cpu
type LevelDBChecker struct {
	db *leveldb.DB
}

func MakeLevelDBChecker(db *leveldb.DB) *LevelDBChecker {
	return &LevelDBChecker{db}
}

func (c *LevelDBChecker) Add(values []interface{}) error {
	for _, val := range values {
		err := c.db.Put([]byte(fmt.Sprint(val)), nil, nil)
		if err != nil {
			log.Print("error with inserting:" + err.Error())
		}
	}
	return nil
}

func (c *LevelDBChecker) Check(values []interface{}) ([]bool, error) {
	if len(values) == 0 {
		return make([]bool, 0), nil
	}
	result := make([]bool, 0, len(values))
	for _, val := range values {
		_, err := c.db.Get([]byte(fmt.Sprint(val)), nil)
		if err == leveldb.ErrNotFound {
			result = append(result, false)
		} else if err != nil {
			log.Print("error with getting:" + err.Error())
		}
		result = append(result, true)
	}
	return result, nil
}
