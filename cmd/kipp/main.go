package main

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"log"
	"mime"
	"net/http"
	"os"
	"time"

	"github.com/6f7262/kipp"
	"github.com/alecthomas/units"
	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type worker time.Duration

func (w worker) Do(ctx context.Context, f func() error) {
loop:
	if err := f(); err != nil {
		log.Fatal(err)
	}
	t := time.After(time.Duration(w))
	select {
	case <-ctx.Done():
		return
	case <-t:
		goto loop
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
	var d kipp.Driver
	s := kipp.Server{Encoding: base64.RawURLEncoding}

	servecmd := kingpin.Command("serve", "Start a kipp server.").Default()

	addr := servecmd.
		Flag("addr", "Server listen address.").
		Default("127.0.0.1:1337").
		String()
	insecure := servecmd.
		Flag("insecure", "Disable https.").
		Bool()
	cert := servecmd.
		Flag("cert", "TLS certificate path.").
		Default("cert.pem").
		String()
	key := servecmd.
		Flag("key", "TLS key path.").
		Default("key.pem").
		String()
	cleanupInterval := servecmd.
		Flag("cleanup-interval", "Cleanup interval for deleting expired files.").
		Default("5m").
		Duration()
	mime := servecmd.
		Flag("mime", "A json formatted collection of extensions and mime types.").
		PlaceHolder("PATH").
		String()
	servecmd.
		Flag("driver", "Available database drivers: mysql, postgres, sqlite3 and mssql.").
		Default("sqlite3").
		StringVar(&d.Dialect)
	servecmd.
		Flag("driver-username", "Database driver username.").
		Default("kipp").
		StringVar(&d.Username)
	servecmd.
		Flag("driver-password", "Database driver password.").
		PlaceHolder("PASSWORD").
		StringVar(&d.Password)
	servecmd.
		Flag("driver-path", "Database driver path. ex: localhost:1337").
		Default("kipp.db").
		StringVar(&d.Path)
	servecmd.
		Flag("expiration", "File expiration time.").
		Default("24h").
		DurationVar(&s.Expiration)
	servecmd.
		Flag("max", "The maximum file size  for uploads.").
		Default("150MB").
		BytesVar((*units.Base2Bytes)(&s.Max))
	servecmd.
		Flag("files", "File path.").
		Default("files").
		StringVar(&s.FilePath)
	servecmd.
		Flag("tmp", "Temp path for in-progress uploads.").
		Default("files/tmp").
		StringVar(&s.TempPath)
	servecmd.
		Flag("public", "Public path for web resources.").
		Default("public").
		StringVar(&s.PublicPath)
	servecmd.
		Flag("proxy-header", "HTTP header to be used for IP logging if set.").
		StringVar(&s.ProxyHeader)

	var u UploadCommand
	{
		uploadcmd := kingpin.Command("upload", "Upload a file.")
		uploadcmd.
			Arg("file", "File to be uploaded").
			Required().
			FileVar(&u.File)
		uploadcmd.
			Flag("insecure", "Don't verify SSL certificates.").
			BoolVar(&u.Insecure)
		uploadcmd.
			Flag("private", "Encrypt the uploaded file").
			BoolVar(&u.Private)
		uploadcmd.
			Flag("url", "Source URL").
			Envar("kipp-upload-url").
			Default("https://kipp.6f.io").
			URLVar(&u.URL)
	}

	statscmd := kingpin.Command("stats", "Display basic stats for the current kipp instance.")
	statscmd.
		Flag("driver", "Available database drivers: mysql, postgres, sqlite3 and mssql.").
		Default("sqlite3").
		StringVar(&d.Dialect)
	statscmd.
		Flag("driver-username", "Database driver username.").
		Default("kipp").
		StringVar(&d.Username)
	statscmd.
		Flag("driver-password", "Database driver password.").
		PlaceHolder("PASSWORD").
		StringVar(&d.Password)
	statscmd.
		Flag("driver-path", "Database driver path. ex: localhost:1337").
		Default("kipp.db").
		StringVar(&d.Path)

	t := kingpin.Parse()

	// kipp upload
	if t == "upload" {
		u.Do()
		return
	}

	// Connect to database
	db, err := d.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	s.DB = db

	// kipp stats - kinda weird, we're doing this after the database opens so
	// we can use it and the upload command is before since it doesn't need it.
	// if t == "stats" {
	// 	var total, size uint64

	// 	t := db.Table("contents")

	// 	t.Select("count(*), sum(size)").Row().Scan(&total, &size)
	// 	fmt.Printf(
	// 		"Total files uploaded: %s (%s)\n",
	// 		humanize.Comma(int64(total)),
	// 		humanize.Bytes(size),
	// 	)

	// 	t.Where("deleted_at IS NULL").Select("count(*), sum(size)").Row().
	// 		Scan(&total, &size)
	// 	fmt.Printf(
	// 		"Currently serving: %s (%s)\n",
	// 		humanize.Comma(int64(total)),
	// 		humanize.Bytes(size),
	// 	)

	// 	t.Where("deleted_at IS NULL").Select("cast(avg(size) as integer)").
	// 		Row().Scan(&size)
	// 	fmt.Printf("Current average file size: %s\n", humanize.Bytes(size))

	// 	t.Select("count(distinct address)").Row().Scan(&total)
	// 	fmt.Printf("Unique IP addresses: %s\n", humanize.Comma(int64(total)))
	// 	return
	// }

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
	if *insecure {
		log.Fatal(hs.ListenAndServe())
	}
	log.Fatal(hs.ListenAndServeTLS(*cert, *key))
}
