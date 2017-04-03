package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/6f7262/conf"
	"github.com/alecthomas/units"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type worker time.Duration

func (w worker) Do(ctx context.Context, f func() error) {
	t := time.After(time.Duration(w))
	for {
		select {
		case <-ctx.Done():
			return
		case <-t:
			// error omitted for now, otherwise log.Fatal
			f()
			t = time.After(time.Duration(w))
		}
	}
}

func main() {
	var s conf.Server
	addr := kingpin.
		Flag("addr", "Server listen address.").
		Default(":1337").
		String()
	cleanupInterval := kingpin.
		Flag("cleanup-interval", "Cleanup interval for deleting expired file").
		Default("5m").
		Duration()
	driver := kingpin.
		Flag("driver", "Available database drivers: mysql, postgres, sqlite3 and mssql").
		Default("sqlite3").
		String()
	driverSource := kingpin.
		Flag("driver-source", "Database driver source. mysql example: user:pass@/database").
		Default("conf.db").
		String()
	kingpin.
		Flag("max", "The maximum file size").
		Default("150MB").
		BytesVar((*units.Base2Bytes)(&s.Max))
	kingpin.
		Flag("file-path", "The path to store uploaded files").
		Default("data/files").
		StringVar(&s.FilePath)
	kingpin.
		Flag("temp-path", "The path to store uploading files").
		Default("data/files/tmp").
		StringVar(&s.TempPath)
	kingpin.
		Flag("public-path", "The path where web resources are located.").
		Default("public").
		StringVar(&s.PublicPath)
	kingpin.Parse()
	db, err := conf.NewDB(*driver, *driverSource)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	s.DB = db
	if err := os.MkdirAll(s.FilePath, 0755); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(s.TempPath, 0755); err != nil {
		log.Fatal(err)
	}
	w := worker(*cleanupInterval)
	go w.Do(context.Background(), s.Cleanup)
	if err := http.ListenAndServe(*addr, s); err != nil {
		log.Fatal(err)
	}
}
