package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"mime"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/uhthomas/kipp"
	"github.com/uhthomas/kipp/internal/httputil"
)

func serve(ctx context.Context) error {
	addr := flag.String("addr", ":80", "listen addr")
	db := flag.String("database", "badger", "database - see docs for more information")
	fs := flag.String("filesystem", "files", "filesystem - see docs for more information")
	web := flag.String("web", "web", "web directory")
	limit := flagBytesValue("limit", 150<<20, "upload limit")
	lifetime := flag.Duration("lifetime", 24*time.Hour, "file lifetime")
	// a negative grace period waits indefinitely
	// a zero grace period immediately terminates
	gracePeriod := flag.Duration("grace-period", time.Minute, "termination grace period")
	flag.Parse()

	for k, v := range mimeTypes {
		for _, vv := range v {
			if err := mime.AddExtensionType(vv, k); err != nil {
				return fmt.Errorf("add mime extension type: %w", err)
			}
		}
	}

	s, err := kipp.New(ctx,
		kipp.ParseDB(*db),
		kipp.ParseFS(*fs),
		kipp.Lifetime(*lifetime),
		kipp.Limit(int64(*limit)),
		kipp.Data(*web),
	)
	if err != nil {
		return err
	}

	log.Printf("listening on %s", *addr)
	return httputil.ListenAndServe(ctx, *addr, s, *gracePeriod)
}
