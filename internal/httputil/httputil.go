package httputil

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	xcontext "github.com/uhthomas/kipp/internal/x/context"
	"golang.org/x/sync/errgroup"
)

func ListenAndServe(ctx context.Context, addr string, h http.Handler, gracePeriod time.Duration) error {
	s := &http.Server{
		Addr: addr,
		// ReadTimeout:  5 * time.Second,
		// WriteTimeout: 10 * time.Second,
		IdleTimeout: 120 * time.Second,
		Handler:     h,
		BaseContext: func(net.Listener) context.Context { return xcontext.Detach(ctx) },
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		<-ctx.Done()
		ctx := ctx
		if gracePeriod != 0 {
			ctx = xcontext.Detach(ctx)
			if gracePeriod > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, gracePeriod)
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
