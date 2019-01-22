package passportChecker

import (
	"database/sql"
	"github.com/pkg/errors"
	"github.com/seiflotfy/cuckoofilter"
	"log"
	"strings"
	"time"
)

//todo: сделать интерфейс Store
//сделать CuckooStore, SQLiteStore
//сделать MultiStore
//CuckooStore может возвращать Maybe return value или NotSureError, тогда MultiStore идёт в SQLiteStore

// возможно переименовать Store в ExistChecker

//попробовать https://github.com/HouzuoGuo/tiedot
//вместо sqlite, возмодно побыстрее
//mariadb точно медленнее(локально на буке)

// на будущее, рассчитывать вероятность в %, задавать необзодимую в MultiStore
// принимать []Store, находить и суммировать вероятности, пока не найдём необходимую
// позволит комбинировать []Store

type Store struct {
	db *sql.DB // очень долго, заливаться будет двое суток
	cf *cuckoo.Filter
}

func NewStore(db *sql.DB, capacity uint) (*Store, error) {
	if err := migrate(db); err != nil {
		return nil, err
	}

	b := make([]byte, 0)
	rows, err := db.Query("SELECT filter FROM cuckoo order by id desc limit 1")
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	var cf *cuckoo.Filter
	if rows.Next() {
		err = rows.Scan(&b)
		if err != nil {
			return nil, err
		}
		cf, err = cuckoo.Decode(b)
		if err != nil {
			return nil, err
		}
	} else {
		cf = cuckoo.NewFilter(capacity)
	}

	return &Store{
		db,
		cf,
	}, nil
}

func (s *Store) Add(strs []string) error {
	toInsert := make([]string, 0)
	for _, str := range strs {
		r, err := s.Check(str)
		if err != nil {
			log.Fatal(err)
			return err
		}
		if !r {
			toInsert = append(toInsert, str)
		}
	}

	q := strings.Builder{}
	q.WriteString("INSERT INTO expired_passports (serie_number, date_added) VALUES ")
	args := make([]interface{}, 0, 2*len(toInsert))
	now := time.Now().Unix()
	for i, str := range toInsert {
		if i == 0 {
			q.WriteString("(?, ?)")
		} else {
			q.WriteString(", (?, ?)")
		}
		args = append(args, str, now)
	}
	stmt, err := s.db.Prepare(q.String())
	if err != nil {
		log.Fatal(err)
		return err
	}
	_, err = stmt.Exec(args...)
	if err != nil {
		log.Fatal(err)
		return err
	}

	for _, str := range toInsert {
		s.cf.Insert([]byte(str))
	}
	return nil
}

func (s *Store) Close() error {
	stmt, _ := s.db.Prepare("INSERT INTO cuckoo (filter) VALUES (?)")
	defer stmt.Close()
	b := s.cf.Encode()
	log.Print(len(b))
	_, err := stmt.Exec(b)
	if err != nil {
		return err
	}
	return s.db.Close()
}

func (s *Store) Check(str string) (bool, error) {
	r := s.cf.Lookup([]byte(str))
	if !r {
		return r, nil
	}
	rows, err := s.db.Query("SELECT count(1) FROM expired_passports WHERE serie_number = \"" + str + "\"")
	if err != nil {
		return true, err
	}
	defer rows.Close()
	if rows.Next() {
		var c int
		if err := rows.Scan(&c); err != nil {
			log.Fatal(c)
			return true, err
		}
		return c != 0, nil
	}
	return true, errors.New("db returns nothing")
}

func migrate(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS cuckoo (id INTEGER PRIMARY KEY, filter BLOB)")
	if err != nil {
		return err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS expired_passports (id INTEGER PRIMARY KEY, serie_number STRING, date_added INTEGER)")
	if err != nil {
		return err
	}
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS date_index ON expired_passports (date_added);")
	if err != nil {
		return err
	}
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS sn_index ON expired_passports (serie_number);")
	if err != nil {
		return err
	}
	return nil
}
