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
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/seiflotfy/cuckoofilter"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
)

const CuckooCapacity = 200000000
const CuckooFilename = "cuckoo.data"

var parseFile = flag.String("parseFile", "", "parse file on start")

//var dbDsn = flag.String("dbDsn", "root:root@tcp(127.0.0.1:3306)/go", "example: root:root@tcp(127.0.0.1:3306)/go")
var port = flag.String("port", "80", "serve port")

//todo: возможно сделать конфигурируемым, где хранить данные
// или просто выбрать лучший вариант и сотавить его

func main() {
	// passwordChecker:passwordChecker@tcp(password-checker.cnqcjxetf5yl.us-east-2.rds.amazonaws.com:3306)/passwordChecker
	flag.Parse()

	logF, err := os.OpenFile("log.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	checkError(err)
	defer logF.Close()
	log.SetOutput(io.MultiWriter(os.Stdout, logF))

	ctx := makeContext()

	//lDb, err := leveldb.OpenFile("dataLevel.db", nil)
	//checkError(err)
	//lCh := passportChecker.MakeLevelDBChecker(lDb)

	//bDb, err := bolt.Open("dataBolt.db", 0600, nil)
	//checkError(err)
	//err = bDb.Update(func(tx *bolt.Tx) error {
	//	_, err := tx.CreateBucketIfNotExists([]byte("data"))
	//	return err
	//})
	//checkError(err)
	//bDbCh := passportChecker.MakeBoltDBChecker(bDb)

	//oh :=  offheap.NewHashTable(130000000)
	//ohC := passportChecker.MakeOffheapChecker(oh)

	//fc := fastcache.New(16 * bytes.GB)
	//fcc := passportChecker.MakeFastCacheChecker(fc)

	//opts := badger.DefaultOptions
	//opts.Dir = "./badger"
	//opts.ValueDir = "./badger"
	//bdb, err := badger.Open(opts)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer bdb.Close()
	//bc := passportChecker.MakeBadgerChecker(bdb)

	//m := passportChecker.MakeMapChecker()

	//pt := passportChecker.MakePrefixTree()
	//err = pt.Add("12")
	//checkError(err)
	//
	//resp2, err := pt.Check("12")
	//checkError(err)
	//log.Fatal(resp2)
	//chPrefix := passportChecker.MakePrefixTreeChecker(pt)

	//db, err := connectDb(*dbDsn)
	//checkError(err)
	//chSql, err := passportChecker.MakeMySQLChecker(db)
	//checkError(err)

	db, err := sql.Open("sqlite3", "./data.sqlite")
	checkError(err)
	chSql, err := passportChecker.MakeSQLiteChecker(db)
	checkError(err)

	//f, err := getCuckoo(CuckooCapacity)
	//checkError(err)

	//chCuckoo, err := passportChecker.MakeCuckooChecker(f)
	//checkError(err)

	//ch := passportChecker.MakeMultiChecker(chCuckoo, chPrefix)

	ch := chSql

	//err = ch.Add([]interface{}{"12"})
	//checkError(err)
	//resp1, err := ch.Check([]interface{}{"12"})
	//checkError(err)
	//log.Fatal(resp1)

	parseDone := make(chan struct{})
	if len(*parseFile) > 0 {
		go func() {
			//defer profile.Start(profile.MemProfile).Stop()
			log.Print("parsing file: " + *parseFile)
			err = AddCSVFile(ctx, ch, *parseFile)
			if err != nil {
				log.Print(err)
				return
			}
			runtime.GC()
			//err = saveCuckoo(f)
			//if err != nil {
			//	log.Print(err)
			//}
			close(parseDone)
		}()
	} else {
		close(parseDone)
	}

	r := chi.NewRouter()
	h := passportChecker.MakeHandler(ch, nil)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/check/{value}", h.Check)
	//r.Get("/count", h.Count)
	//r.Get("/get-from/{ts}", h.GetFrom)

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

	<-parseDone
}

// todo: переделать в сервис. Типа Importer
// сделать handler для него, со прогрессом и тд
// что бы он ещё смог сам скачивать и импортить по запросу
// предусмотреть докачку при обрыве, с количеством попыток
func AddCSVFile(ctx context.Context, ch passportChecker.ExistChecker, path string) error {
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
		p, err := passportChecker.MakePassport(
			strings.Replace(strings.Replace(line[0], " ", "", -1), "-", "", -1),
			strings.Replace(strings.Replace(line[1], " ", "", -1), "-", "", -1))
		if err != nil {
			log.Println(err.Error())
			continue
		}
		chunk = append(chunk, p.Uint64())
		if len(chunk) == cap(chunk) {
			err = ch.Add(chunk)
			if err != nil {
				return err
			}
			chunk = make([]interface{}, 0, cap(chunk))
		}
		readCount++
		if readCount%1000000 == 0.0 {
			log.Printf("read: %v", readCount)
		}
		//if readCount%10000000 == 0.0 {
		//	runtime.GC()
		//}
	}
}

func getCuckoo(cap uint) (*cuckoo.Filter, error) {
	b, _ := ioutil.ReadFile("./" + CuckooFilename)
	if len(b) == 0 {
		return cuckoo.NewFilter(cap), nil
	}
	return cuckoo.Decode(b)
}

func saveCuckoo(f *cuckoo.Filter) error {
	log.Print("saving Cuckoo..")
	err := ioutil.WriteFile("./"+CuckooFilename, f.Encode(), 0644)
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

func connectDb(dbDsn string) (*sql.DB, error) {
	con, err := sql.Open("mysql", dbDsn)
	if err != nil {
		return nil, err
	}
	if err := con.Ping(); err != nil {
		return nil, err
	}
	return con, nil
}
