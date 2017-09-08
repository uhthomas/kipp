package main

import (
	"context"
	"crypto/tls"
	"encoding/base32"
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
			if err := f(); err != nil {
				log.Fatal(err)
			}
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
	var d conf.Driver
	s := conf.Server{
		Encoding: base32.NewEncoding("0123456789abcdefghjkmnpqrtuvwxyz").
			WithPadding(base32.NoPadding),
	}
	addr := kingpin.
		Flag("addr", "Server listen address.").
		Default(":1337").
		String()
	secure := kingpin.
		Flag("secure", "Enable https.").
		Bool()
	cert := kingpin.
		Flag("cert", "TLS certificate path.").
		Default("cert.pem").
		String()
	key := kingpin.
		Flag("key", "TLS key path.").
		Default("key.pem").
		String()
	cleanupInterval := kingpin.
		Flag("cleanup-interval", "Cleanup interval for deleting expired files.").
		Default("5m").
		Duration()
	mime := kingpin.
		Flag("mime", "A json formatted collection of extensions and mime types.").
		PlaceHolder("PATH").
		String()
	kingpin.
		Flag("driver", "Available database drivers: mysql, postgres, sqlite3 and mssql.").
		Default("sqlite3").
		StringVar(&d.Dialect)
	kingpin.
		Flag("driver-username", "Database driver username.").
		Default("conf").
		StringVar(&d.Username)
	kingpin.
		Flag("driver-password", "Database driver password.").
		PlaceHolder("PASSWORD").
		StringVar(&d.Password)
	kingpin.
		Flag("driver-path", "Database driver path. ex: localhost:1337").
		Default("conf.db").
		StringVar(&d.Path)
	kingpin.
		Flag("expiration", "File expiration time.").
		Default("24h").
		DurationVar(&s.Expiration)
	kingpin.
		Flag("max", "The maximum file size  for uploads.").
		Default("150MB").
		BytesVar((*units.Base2Bytes)(&s.Max))
	kingpin.
		Flag("files", "File path.").
		Default("files").
		StringVar(&s.FilePath)
	kingpin.
		Flag("tmp", "Temp path for in-progress uploads.").
		Default("files/tmp").
		StringVar(&s.TempPath)
	kingpin.
		Flag("public", "Public path for web resources.").
		Default("public").
		StringVar(&s.PublicPath)
	kingpin.Parse()

	// Load mime types
	if m := *mime; m != "" {
		if err := loadMimeTypes(m); err != nil {
			log.Fatal(err)
		}
	}

	// Make paths for files and temp files
	if err := os.MkdirAll(s.FilePath, 0755); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(s.TempPath, 0755); err != nil {
		log.Fatal(err)
	}

	// Connect to database
	db, err := d.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	s.DB = db

	// Start cleanup worker
	if s.Expiration > 0 {
		w := worker(*cleanupInterval)
		go w.Do(context.Background(), s.Cleanup)
	}

	// Output a message so users know when the server has been started.
	log.Printf("Listening on %s", *addr)

	// Start HTTP server
	hs := &http.Server{
		Addr:    *addr,
		Handler: s,
		TLSConfig: &tls.Config{
			PreferServerCipherSuites: true,
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.X25519,
			},
			MinVersion: tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		},
		// ReadTimeout:  5 * time.Second,
		// WriteTimeout: 10 * time.Second,
		IdleTimeout: 120 * time.Second,
	}
	if *secure {
		log.Fatal(hs.ListenAndServeTLS(*cert, *key))
	}
	log.Fatal(hs.ListenAndServe())
}
