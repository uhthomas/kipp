package context

import (
	"context"
	"time"
)

type detachedContext struct{ context.Context }

func Detach(ctx context.Context) context.Context { return detachedContext{Context: ctx} }

func (detachedContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (detachedContext) Done() <-chan struct{}       { return nil }
func (detachedContext) Err() error                  { return nil }
