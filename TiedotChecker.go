package passportChecker

import (
	tiedot "github.com/HouzuoGuo/tiedot/db"
	"time"
)

const CollectionKey = "values"

type TiedotChecker struct {
	db *tiedot.DB
}

func MakeTiedotChecker(db *tiedot.DB) (*TiedotChecker, error) {
	tc := &TiedotChecker{db: db}
	return tc, tc.migrate()
}

func (tc *TiedotChecker) Add(values []interface{}) error {
	coll := tc.collection()
	now := time.Now().Unix()
	for _, val := range values {
		_, err := coll.Insert(map[string]interface{}{
			"value": val,
			"ts":    now})
		if err != nil {
			return err
		}
	}
	inCh := make(chan interface{})
	go func() {
		for _, val := range values {
			inCh <- val
		}
		close(inCh)
	}()
	errCh := make(chan error)
	for i := 0; i < 4; i++ {
		go func() {
			coll := tc.collection()
			for val := range inCh {
				_, err := coll.Insert(map[string]interface{}{
					"value": val,
					"ts":    now})
				if err != nil {
					errCh <- err
					return
				}
			}
			errCh <- nil
		}()
	}
	errs := make([]error, 0)
	for i := 0; i < 4; i++ {
		err := <-errCh
		if err != nil {
			errs = append(errs, err)
		}
	}
	for _, err := range errs {
		return err
	}
	return nil
}

func (tc *TiedotChecker) Check(values []interface{}) ([]bool, error) {
	result := make([]bool, len(values))
	coll := tc.collection()
	for i, val := range values {
		query := map[string]interface{}{
			"eq":    val,
			"in":    []interface{}{"value"},
			"limit": 1,
		}
		queryResult := make(map[int]struct{})
		if err := tiedot.EvalQuery(query, coll, &queryResult); nil != err {
			return result, err
		}
		result[i] = len(queryResult) > 0
	}
	return result, nil
}

func (tc *TiedotChecker) migrate() error {
	collection := tc.db.Use(CollectionKey)
	if collection == nil {
		if err := tc.db.Create(CollectionKey); err != nil {
			return err
		}
		collection = tc.db.Use(CollectionKey)
	}
	_ = collection.Index([]string{"value"})
	_ = collection.Index([]string{"ts"})
	return nil
}

func (tc *TiedotChecker) collection() *tiedot.Col {
	return tc.db.Use(CollectionKey)
}
