package main

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"githib.com/coraxster/passportChecker"
	"github.com/labstack/gommon/log"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"os"
)

func main() {
	db, err := sql.Open("sqlite3", "./test2.db")
	checkError(err)
	s, err := passportChecker.NewStore(db, 500000000)
	defer s.Close()
	checkError(err)

	//lc := 118 544 508
	file, err := os.Open("/Users/dmitry.kuzmin/dev/test/passports/list_of_expired_passports.csv")
	checkError(err)
	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))
	_, err = reader.Read() // skip header line
	checkError(err)
	chunk := make([]string, 0, 300)
	var readCount uint
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}
		checkError(err)
		chunk = append(chunk, line[0]+line[1])
		if len(chunk) == cap(chunk) {
			err = s.Add(chunk)
			checkError(err)
			chunk = make([]string, 0, cap(chunk))
		}
		readCount++
		if readCount%100000 == 0.0 {
			log.Print(readCount)
		}
	}
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
