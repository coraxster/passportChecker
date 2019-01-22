package main

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/csv"
	"githib.com/coraxster/passportChecker"
	"github.com/labstack/gommon/log"
	_ "github.com/mattn/go-sqlite3"
	"github.com/seiflotfy/cuckoofilter"
	"io"
	"os"
	"os/signal"
)

const CuckooCapacity = 200000000

func main() {
	ctx := makeContext()
	db, err := sql.Open("sqlite3", "./test.db")
	checkError(err)

	f, err := getCuckoo(db, CuckooCapacity)
	checkError(err)

	chSql, err := passportChecker.MakeSQLiteChecker(db)
	checkError(err)

	chCuckoo, err := passportChecker.MakeCuckooChecker(f)
	checkError(err)

	ch := passportChecker.MakeMultiChecker(chCuckoo, chSql)

	AddCSVFile(ctx, ch, "/Users/dmitry.kuzmin/dev/test/passports/list_of_expired_passports.csv")

	err = saveCuckoo(db, f)
	checkError(err)
}

func AddCSVFile(ctx context.Context, ch *passportChecker.MultiChecker, path string) {
	//lc := 118 544 508
	file, err := os.Open(path)
	checkError(err)
	reader := csv.NewReader(bufio.NewReader(file))
	_, err = reader.Read() // skip header line
	checkError(err)
	chunk := make([]interface{}, 0, 100000)
	var readCount uint
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		line, err := reader.Read()
		if err == io.EOF {
			return
		}
		checkError(err)
		chunk = append(chunk, line[0]+line[1])
		if len(chunk) == cap(chunk) {
			err = ch.Add(chunk)
			checkError(err)
			chunk = make([]interface{}, 0, cap(chunk))
		}
		readCount++
		if readCount%100000 == 0.0 {
			log.Print(readCount)
		}
	}
}

func getCuckoo(db *sql.DB, cap uint) (*cuckoo.Filter, error) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS cuckoo_store (id INTEGER PRIMARY KEY, filter BLOB)")
	if err != nil {
		return nil, err
	}
	row := db.QueryRow("SELECT filter FROM cuckoo_store order by id desc limit 1")
	b := make([]byte, 0)
	err = row.Scan(&b)
	if err != sql.ErrNoRows && err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return cuckoo.NewFilter(cap), nil
	}
	return cuckoo.Decode(b)
}

func saveCuckoo(db *sql.DB, f *cuckoo.Filter) error {
	log.Print("saving Cuckoo..")
	_, err := db.Exec("INSERT INTO cuckoo_store (filter) VALUES (?)", f.Encode())
	return err
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func makeContext() context.Context {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		cancel()
		signal.Stop(c)
	}()
	return ctx
}
