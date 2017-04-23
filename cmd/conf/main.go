package main

import (
	"context"
	"encoding/json"
	"log"
	"mime"
	"net/http"
	"os"
	"time"

	"github.com/6f7262/conf"
	"github.com/alecthomas/units"
	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
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

func loadMimeTypes(path string) error {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	m := make(map[string][]string)
	if err := json.NewDecoder(f).Decode(&m); err != nil {
		return err
	}
	for k, v := range m {
		for _, vv := range v {
			mime.AddExtensionType(vv, k)
		}
	}
	return nil
}

func main() {
	var s conf.Server
	addr := kingpin.
		Flag("addr", "Server listen address.").
		Default(":1337").
		String()
	cleanupInterval := kingpin.
		Flag("cleanup-interval", "Cleanup interval for deleting expired files.").
		Default("5m").
		Duration()
	driver := kingpin.
		Flag("driver", "Available database drivers: mysql, postgres, sqlite3 and mssql.").
		Default("sqlite3").
		String()
	driverSource := kingpin.
		Flag("driver-source", "Database driver source. mysql example: user:pass@/database.").
		Default("conf.db").
		String()
	mime := kingpin.
		Flag("mime", "A json formatted collection of extensions and mime types.").
		String()
	kingpin.
		Flag("expiration", "File expiration time.").
		Default("24h").
		DurationVar(&s.Expiration)
	kingpin.
		Flag("max", "The maximum file size limit for uploads.").
		Default("150MB").
		BytesVar((*units.Base2Bytes)(&s.Max))
	kingpin.
		Flag("file-path", "The path to store uploaded files.").
		Default("files").
		StringVar(&s.FilePath)
	kingpin.
		Flag("temp-path", "The path to store uploading files.").
		Default("files/tmp").
		StringVar(&s.TempPath)
	kingpin.
		Flag("public-path", "The path where web resources are located.").
		Default("public").
		StringVar(&s.PublicPath)
	kingpin.Parse()

	// Load mime types
	if m := *mime; m != "" {
		if err := loadMimeTypes(m); err != nil {
			log.Fatal(err)
		}
	}

	// Load database
	db, err := conf.NewDB(*driver, *driverSource)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	s.DB = db

	// Make paths for files and temp files
	if err := os.MkdirAll(s.FilePath, 0755); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(s.TempPath, 0755); err != nil {
		log.Fatal(err)
	}

	// Start cleanup worker
	if s.Expiration > 0 {
		w := worker(*cleanupInterval)
		go w.Do(context.Background(), s.Cleanup)
	}

	// Output a message so users know when the server has been started.
	log.Printf("Listening on %s", *addr)

	// Start HTTP server
	if err := http.ListenAndServe(*addr, s); err != nil {
		log.Fatal(err)
	}
}
