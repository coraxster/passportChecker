package passportChecker

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const ChunkSize = 300

type SQLiteChecker struct {
	db *sql.DB
}

func MakeSQLiteChecker(db *sql.DB) (*SQLiteChecker, error) {
	if err := migrate(db); err != nil {
		return nil, err
	}
	return &SQLiteChecker{db}, nil
}

func (c *SQLiteChecker) Add(values []interface{}) error {
	if len(values) == 0 {
		return nil
	}
	var err error
	var i int
	for i = 0; i+ChunkSize < len(values); i = i + ChunkSize {
		err = c.addChunk(values[i : i+ChunkSize])
		if err != nil {
			return err
		}
	}
	if i < len(values) {
		return c.addChunk(values[i:])
	}
	return nil
}

func (c *SQLiteChecker) addChunk(values []interface{}) error {
	q := strings.Builder{}
	q.WriteString("INSERT OR IGNORE INTO value_store (val, date_added) VALUES ")
	args := make([]interface{}, 0, 2*len(values))
	now := time.Now().Unix()
	for i, val := range values {
		if i == 0 {
			q.WriteString("(?, ?)")
		} else {
			q.WriteString(", (?, ?)")
		}
		args = append(args, fmt.Sprint(val), now)
	}
	_, err := c.db.Exec(q.String(), args...)
	if err != nil {
		return err
	}
	return nil
}

func (c *SQLiteChecker) Count() (int, error) {
	row := c.db.QueryRow("SELECT COALESCE(MAX(id)+1, 0) FROM value_store")
	var i int
	err := row.Scan(&i)
	return i, err
}

func (c *SQLiteChecker) Check(values []interface{}) ([]bool, error) {
	result := make([]bool, len(values))
	if len(values) == 0 {
		return result, nil
	}
	var i int
	for i = 0; i+ChunkSize < len(values); i = i + ChunkSize {
		r, err := c.checkChunk(values[i : i+ChunkSize])
		if err != nil {
			return result, err
		}
		result = append(result, r...)
	}
	if i < len(values) {
		r, err := c.checkChunk(values[i:])
		if err != nil {
			return result, err
		}
		result = append(result, r...)
	}
	return result, nil
}

func (c *SQLiteChecker) checkChunk(values []interface{}) ([]bool, error) {
	strs := make([]string, 0, len(values))
	strMap := make(map[string]bool)
	q := strings.Builder{}
	q.WriteString("SELECT val, count(1) FROM value_store WHERE val IN (\"")
	for i, val := range values {
		s := fmt.Sprint(val)
		strs = append(strs, s)
		strMap[s] = false
		if i == 0 {
			q.WriteString(fmt.Sprint(val))
			q.WriteString("\"")
		} else {
			q.WriteString(", \"")
			q.WriteString(s)
			q.WriteString("\"")
		}
	}
	q.WriteString(") GROUP BY val")
	rows, err := c.db.Query(q.String())
	if err != nil {
		return []bool{}, err
	}
	for rows.Next() {
		var v string
		var c int
		if err := rows.Scan(&v, &c); err != nil {
			return []bool{}, err
		}
		if c != 0 {
			strMap[v] = true
		}
	}
	result := make([]bool, len(values))
	for i, s := range strs {
		result[i] = strMap[s]
	}
	return result, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS value_store (id INTEGER PRIMARY KEY, val TEXT UNIQUE, date_added INTEGER)")
	if err != nil {
		return err
	}
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS date_index ON value_store (date_added);")
	if err != nil {
		return err
	}
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS sn_index ON value_store (val);")
	if err != nil {
		return err
	}
	return nil
}
