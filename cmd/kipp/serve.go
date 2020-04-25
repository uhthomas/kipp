package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/uhthomas/kipp/database/sqlite3"
	"github.com/uhthomas/kipp/filesystem/local"

	"github.com/uhthomas/kipp"
)

func serve(ctx context.Context) error {
	addr := flag.String("addr", ":80", "listen addr")
	dsn := flag.String("dsn", "file::memory:?cache=shared", "data source name")
	dir := flag.String("dir", "files", "file directory")
	tmp := flag.String("tmp", os.TempDir(), "tmp directory")
	limit := flagBytesValue("limit", 150<<20, "upload limit")
	lifetime := flag.Duration("lifetime", 24*time.Hour, "file lifetime")
	flag.Parse()

	for k, v := range mimeTypes {
		for _, vv := range v {
			if err := mime.AddExtensionType(vv, k); err != nil {
				return fmt.Errorf("add mime extension type: %w", err)
			}
		}
	}

	fs, err := local.New(*dir, *tmp)
	if err != nil {
		return fmt.Errorf("new local filesystem: %w", err)
	}

	db, err := sqlite3.New(ctx, *dsn)
	if err != nil {
		return fmt.Errorf("new sqlite3 database: %w", err)
	}
	defer db.Close(ctx)

	log.Printf("listening on %s", *addr)

	return (&http.Server{
		Addr: *addr,
		Handler: &kipp.Server{
			Database:   db,
			FileSystem: fs,
			Limit:      int64(*limit),
			Lifetime:   *lifetime,
			PublicPath: "some public path",
		},
		// ReadTimeout:  5 * time.Second,
		// WriteTimeout: 10 * time.Second,
		BaseContext: func(net.Listener) context.Context { return ctx },
	}).ListenAndServe()
}
