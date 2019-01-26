package passportChecker

import (
	"fmt"
	"github.com/dgraph-io/badger"
	"log"
)

type BadgerChecker struct {
	db *badger.DB
}

//10gb, high cpu, 1-5ms on true resolve
func MakeBadgerChecker(db *badger.DB) *BadgerChecker {
	return &BadgerChecker{db}
}

func (c *BadgerChecker) Add(values []interface{}) error {
	err := c.db.Update(func(txn *badger.Txn) error {
		for _, val := range values {
			err := txn.Set([]byte(fmt.Sprint(val)), nil)
			if err != nil {
				log.Print(err.Error())
				continue
			}
		}
		return nil
	})
	return err
}

func (c *BadgerChecker) Check(values []interface{}) ([]bool, error) {
	if len(values) == 0 {
		return make([]bool, 0), nil
	}
	result := make([]bool, 0, len(values))
	err := c.db.View(func(txn *badger.Txn) error {
		for _, val := range values {
			_, err := txn.Get([]byte(fmt.Sprint(val)))
			if err == badger.ErrKeyNotFound {
				result = append(result, false)
				continue
			}
			if err != nil {
				return err
			}
			result = append(result, true)
		}
		return nil
	})
	return result, err
}
