package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"mime"
	"net"
	"net/http"
	"time"

	_ "github.com/lib/pq"
	"github.com/uhthomas/kipp"
	"github.com/uhthomas/kipp/internal/databaseutil"
	"github.com/uhthomas/kipp/internal/filesystemutil"
	xcontext "github.com/uhthomas/kipp/internal/x/context"
	"golang.org/x/sync/errgroup"
)

func serve(ctx context.Context) error {
	addr := flag.String("addr", ":80", "listen addr")
	dbf := flag.String("database", "badger", "database - see docs for more information")
	fsf := flag.String("filesystem", "files", "filesystem - see docs for more information")
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

	fs, err := filesystemutil.Parse(*fsf)
	if err != nil {
		return fmt.Errorf("parse filesystem: %w", err)
	}

	db, err := databaseutil.Parse(ctx, *dbf)
	if err != nil {
		return fmt.Errorf("parse database: %w", err)
	}
	defer db.Close(ctx)

	log.Printf("listening on %s", *addr)

	g, ctx := errgroup.WithContext(ctx)

	s := &http.Server{
		Addr: *addr,
		Handler: &kipp.Server{
			Database:   db,
			FileSystem: fs,
			Limit:      int64(*limit),
			Lifetime:   *lifetime,
			PublicPath: *web,
		},
		// ReadTimeout:  5 * time.Second,
		// WriteTimeout: 10 * time.Second,
		BaseContext: func(net.Listener) context.Context { return xcontext.Detach(ctx) },
	}

	g.Go(func() error {
		<-ctx.Done()
		ctx := ctx
		if *gracePeriod != 0 {
			ctx = xcontext.Detach(ctx)
			if *gracePeriod > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, *gracePeriod)
				defer cancel()
			}
		}
		return s.Shutdown(ctx)
	})

	g.Go(func() error {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen and serve: %w", err)
		}
		return nil
	})

	return g.Wait()
}
