package passportChecker

import (
	"github.com/dgraph-io/badger"
	"log"
)

type BadgerChecker struct {
	db *badger.DB
}

//3.5gb ram, 9.5gb disk, high 1 cpu,
//2gb ram, 8gb disk, high 1 cpu,
//2gb ram, 7.7gb disk, high 1 cpu, записывая uint64 вместо строчек, самый замороченный вариант
//2gb ram, 8.1gb disk, high 1 cpu, тупо строчки, то что нужно

func MakeBadgerChecker(db *badger.DB) *BadgerChecker {
	return &BadgerChecker{db}
}

func (c *BadgerChecker) Add(values []string) error {
	err := c.db.Update(func(txn *badger.Txn) error {
		for _, val := range values {
			bs := []byte(val)
			_, err := txn.Get(bs)
			if err == badger.ErrKeyNotFound {
				err = txn.Set(bs, nil)
			}
			if err != nil {
				log.Print(err.Error())
				continue
			}
		}
		return nil
	})
	return err
}

func (c *BadgerChecker) Check(values []string) ([]bool, error) {
	if len(values) == 0 {
		return make([]bool, 0), nil
	}
	result := make([]bool, 0, len(values))
	err := c.db.View(func(txn *badger.Txn) error {
		for _, val := range values {
			bs := []byte(val)
			_, err := txn.Get(bs)
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
