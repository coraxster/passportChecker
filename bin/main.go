package main

import (
	"bufio"
	"compress/bzip2"
	"context"
	"flag"
	"github.com/coraxster/passportChecker"
	"github.com/dgraph-io/badger"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/mileusna/crontab"
	"github.com/pkg/errors"
	"github.com/seiflotfy/cuckoofilter"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
)

const CuckooCapacity = 200000000

var url = flag.String("url", "https://guvm.mvd.ru/upload/expired-passports/list_of_expired_passports.csv.bz2", "passport numbers bz2 url")
var port = flag.String("port", "80", "serve port")
var updateCron = flag.String("updateCron", "0 2 * * *", "when update")
var updateOnStart = flag.Bool("updateOnStart", false, "update on start")

var parseMutex sync.Mutex

func main() {
	flag.Parse()

	logFile, err := os.OpenFile("log.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	checkError(err)
	defer logFile.Close()
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	middleware.DefaultLogger = middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: log.New(io.MultiWriter(os.Stdout, logFile), "", log.LstdFlags), NoColor: false})

	ctx := makeContext()

	bdb := getBadger()
	defer bdb.Close()
	badCh := passportChecker.MakeBadgerChecker(bdb)

	filter, err := getCuckoo(bdb)
	checkError(err)
	log.Println("INFO: Cuckoo size: ", filter.Count())
	CuCh, err := passportChecker.MakeCuckooChecker(filter)
	checkError(err)

	ch := passportChecker.MakeMultiChecker(CuCh, badCh)

	updateFunc := func() {
		parseMutex.Lock()
		defer parseMutex.Unlock()
		changed, err := updateChecker(ctx, ch, bdb, *url)
		if err != nil {
			log.Print("ERROR: while updating. ", err.Error())
			return
		}
		if changed {
			saveCuckoo(bdb, filter)
		}
	}

	if *updateOnStart {
		go func() {
			updateFunc()
		}()
	}
	log.Println("INFO: Crontab config: ", *updateCron)
	ctab := crontab.New()
	err = ctab.AddJob(*updateCron, updateFunc)
	checkError(err)

	err = startServer(ctx, ch)
	checkError(err)

	parseMutex.Lock()
}

func startServer(ctx context.Context, ch *passportChecker.MultiChecker) error {
	r := chi.NewRouter()
	h := passportChecker.MakeHandler(ch)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/check/{value}", h.Check)
	srv := http.Server{Addr: ":" + *port, Handler: r}
	log.Print("INFO: starting serving on :" + *port)
	go func() {
		err := srv.ListenAndServe()
		if err != http.ErrServerClosed {
			checkError(err)
		}
	}()
	<-ctx.Done()
	return srv.Shutdown(context.Background())
}

func getBadger() *badger.DB {
	opts := badger.DefaultOptions
	opts.Dir = "./badger"
	opts.ValueDir = "./badger"
	opts.SyncWrites = false
	opts.ValueLogMaxEntries = 100000000
	bdb, err := badger.Open(opts)
	checkError(err)
	return bdb
}

func updateChecker(ctx context.Context, ch passportChecker.ExistChecker, bdb *badger.DB, url string) (bool, error) {
	log.Print("INFO: parsing url: " + url)
	localVersion, err := getLocalVersion(bdb)
	if err != nil {
		return false, err
	}
	log.Println("INFO: Local bz2 version: ", localVersion)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Add("If-None-Match", localVersion)
	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return false, err
	}
	if resp.StatusCode == 304 {
		log.Println("INFO: Remote version has not changed")
		return false, nil
	}
	remoteVersion := resp.Header.Get("ETag")
	if remoteVersion == "" {
		return false, errors.New("Server returns response without ETag header")
	}
	log.Println("INFO: Remote bz2 version: ", remoteVersion)
	reader := bufio.NewReader(bzip2.NewReader(resp.Body))
	err = fillChecker(ctx, ch, reader)
	if err != nil {
		return false, err
	}
	return true, setLocalVersion(bdb, remoteVersion)
}

func fillChecker(ctx context.Context, ch passportChecker.ExistChecker, reader *bufio.Reader) error {
	log.Println("INFO: Start filling ")
	chunkCh := make(chan []string)
	defer close(chunkCh)
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for chunk := range chunkCh {
				err := ch.Add(chunk)
				if err != nil {
					log.Printf("ERROR: Add error: %v", err)
				}
			}
		}()
	}
	chunk := make([]string, 0, 1000)
	var readCount uint
	var err error
forLoop:
	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			break forLoop
		default:
		}
		var line []byte
		line, _, err = reader.ReadLine()
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
			log.Printf("INFO: Read: %v", readCount)
		}
	}
	log.Printf("INFO: Read done: %v", readCount)
	return err
}

func getLocalVersion(bdb *badger.DB) (string, error) {
	var localVersion string
	err := bdb.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("bzVersion"))
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		err = item.Value(func(v []byte) error {
			localVersion = string(v)
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	return localVersion, err
}

func setLocalVersion(bdb *badger.DB, localVersion string) error {
	return bdb.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("bzVersion"), []byte(localVersion))
	})
}

func getCuckoo(db *badger.DB) (*cuckoo.Filter, error) {
	log.Print("INFO: reading Cuckoo..")
	b := make([]byte, 0)
	err := db.View(func(txn *badger.Txn) error {
		i, err := txn.Get([]byte("Cuckoo"))
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		b, err = i.ValueCopy(b)
		return err
	})
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return cuckoo.NewFilter(CuckooCapacity), nil
	}
	return cuckoo.Decode(b)
}

func saveCuckoo(db *badger.DB, f *cuckoo.Filter) {
	log.Print("INFO: saving Cuckoo..")
	err := db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("Cuckoo"), f.Encode())
	})
	if err != nil {
		log.Printf("ERROR: saving Cuckoo failed: %v", err.Error())
	}
	log.Print("INFO: cuckoo saved")
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
