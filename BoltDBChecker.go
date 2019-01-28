package passportChecker

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/labstack/gommon/log"
)

// медленно пишет, 100% cpu
type BoltDBChecker struct {
	db *bolt.DB
}

// super slow :(
func MakeBoltDBChecker(db *bolt.DB) *BoltDBChecker {
	return &BoltDBChecker{db}
}

func (c *BoltDBChecker) Add(values []interface{}) error {
	err := c.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("data"))
		for _, val := range values {
			err := bucket.Put([]byte(fmt.Sprint(val)), []byte(""))
			if err != nil {
				log.Print("error with inserting:" + err.Error())
			}
		}
		return nil
	})
	if err != nil {
		log.Print("error with inserting:" + err.Error())
	}
	return nil
}

func (c *BoltDBChecker) Check(values []interface{}) ([]bool, error) {
	if len(values) == 0 {
		return make([]bool, 0), nil
	}
	result := make([]bool, 0, len(values))
	err := c.db.View(func(tx *bolt.Tx) error {
		for _, val := range values {
			bucket := tx.Bucket([]byte("data"))
			r := bucket.Get([]byte(fmt.Sprint(val)))
			result = append(result, r != nil)
		}
		return nil
	})
	if err != nil {
		log.Print("error with getting:" + err.Error())
	}
	return result, nil
}
