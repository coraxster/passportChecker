package main

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/csv"
	"flag"
	"github.com/coraxster/passportChecker"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/mattn/go-sqlite3"
	"github.com/seiflotfy/cuckoofilter"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
)

const CuckooCapacity = 200000000
const StateFilename = "state.sql"

var parseFile = flag.String("parseFile", "", "parse file on start")
var port = flag.String("port", "80", "serve port")

func main() {
	flag.Parse()
	ctx := makeContext()
	db, err := sql.Open("sqlite3", "./"+StateFilename)
	checkError(err)

	f, err := getCuckoo(db, CuckooCapacity)
	checkError(err)

	chSql, err := passportChecker.MakeSQLiteChecker(db)
	checkError(err)

	chCuckoo, err := passportChecker.MakeCuckooChecker(f)
	checkError(err)

	ch := passportChecker.MakeMultiChecker(chCuckoo, chSql)

	go func() {
		if len(*parseFile) > 0 {
			log.Print("parsing file: " + *parseFile)
			err = AddCSVFile(ctx, ch, *parseFile)
			if err != nil {
				log.Print(err)
			}
			err = saveCuckoo(db, f)
			if err != nil {
				log.Print(err)
			}
		}
	}()

	h := passportChecker.MakeHandler(ch, chSql)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/check/{value}", h.Check)
	r.Get("/count", h.Count)

	srv := http.Server{Addr: ":" + *port, Handler: r}
	log.Print("starting serving on :" + *port)
	go func() {
		err := srv.ListenAndServe()
		if err != http.ErrServerClosed {
			checkError(err)
		}
	}()
	<-ctx.Done()
	err = srv.Shutdown(context.Background())
	checkError(err)
}

func AddCSVFile(ctx context.Context, ch *passportChecker.MultiChecker, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	reader := csv.NewReader(bufio.NewReader(file))
	_, err = reader.Read() // skip header line
	if err != nil {
		return err
	}
	chunk := make([]interface{}, 0, 1000)
	var readCount uint
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		line, err := reader.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		chunk = append(chunk, line[0]+line[1])
		if len(chunk) == cap(chunk) {
			err = ch.Add(chunk)
			if err != nil {
				return err
			}
			chunk = make([]interface{}, 0, cap(chunk))
		}
		readCount++
		if readCount%100000 == 0.0 {
			log.Printf("read: %v", readCount)
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
	log.Print("saving Cuckoo...")
	_, err := db.Exec("INSERT INTO cuckoo_store (filter) VALUES (?)", f.Encode())
	log.Print("cuckoo saved")
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
