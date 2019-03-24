package main

import (
	"bufio"
	"context"
	"flag"
	"github.com/coraxster/passportChecker"
	"github.com/dgraph-io/badger"
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
)

const CuckooCapacity = 200000000
const CuckooFilename = "cuckoo.data"

var parseFile = flag.String("parseFile", "", "parse file on start")
var port = flag.String("port", "80", "serve port")

func main() {
	flag.Parse()

	logFile, err := os.OpenFile("log.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	checkError(err)
	defer logFile.Close()
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	middleware.DefaultLogger = middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: log.New(io.MultiWriter(os.Stdout, logFile), "", log.LstdFlags), NoColor: false})

	ctx := makeContext()

	opts := badger.DefaultOptions
	opts.Dir = "./badger"
	opts.ValueDir = "./badger"
	opts.SyncWrites = false
	bdb, err := badger.Open(opts)
	checkError(err)
	defer bdb.Close()
	badCh := passportChecker.MakeBadgerChecker(bdb)

	f, err := getCuckoo(CuckooCapacity)
	log.Println("Cuckoo size: ", f.Count())
	checkError(err)
	CuCh, err := passportChecker.MakeCuckooChecker(f)
	checkError(err)

	ch := passportChecker.MakeMultiChecker(CuCh, badCh)

	parseDone := make(chan struct{})
	if len(*parseFile) > 0 {
		go func() {
			log.Print("parsing file: " + *parseFile)
			err = AddCSVFile(ctx, ch, *parseFile)
			if err != nil {
				log.Print(err)
				return
			}
			runtime.GC()
			saveCuckoo(f)
			close(parseDone)
		}()
	} else {
		close(parseDone)
	}

	r := chi.NewRouter()
	h := passportChecker.MakeHandler(ch)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/check/{value}", h.Check)

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

func AddCSVFile(ctx context.Context, ch passportChecker.ExistChecker, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	reader := bufio.NewReader(file)
	_, _, err = reader.ReadLine() // skip header line
	if err != nil {
		return err
	}

	chunkCh := make(chan []string)
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for chunk := range chunkCh {
				err = ch.Add(chunk)
				if err != nil {
					close(chunkCh)
					log.Printf("Add error: %v", err)
				}
			}
		}()
	}
	chunk := make([]string, 0, 1000)
	var readCount uint
	for {
		select {
		case <-ctx.Done():
			break
		default:
		}
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			break
		}
		chunk = append(chunk, string(line))
		if len(chunk) == cap(chunk) {
			chunkCh <- chunk
			chunk = make([]string, 0, cap(chunk))
		}
		readCount++
		if readCount%1000000 == 0.0 {
			log.Printf("read: %v", readCount)
		}
	}
	close(chunkCh)
	return err
}

func getCuckoo(cap uint) (*cuckoo.Filter, error) {
	b, _ := ioutil.ReadFile("./" + CuckooFilename)
	if len(b) == 0 {
		return cuckoo.NewFilter(cap), nil
	}
	return cuckoo.Decode(b)
}

func saveCuckoo(f *cuckoo.Filter) {
	log.Print("saving Cuckoo..")
	err := ioutil.WriteFile("./"+CuckooFilename, f.Encode(), 0644)
	if err != nil {
		log.Printf("rerror while saving Cuckoo: %v", err.Error())
	}
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
