package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

type worker time.Duration

func (w worker) Do(ctx context.Context, f func() error) {
	for {
		if err := f(); err != nil {
			log.Fatal(err)
		}
		t := time.NewTimer(time.Duration(w))
		select {
		case <-t.C:
		case <-ctx.Done():
			t.Stop()
			return
		}
	}
}

func Main(ctx context.Context) error {
	var cmd string
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "", "serve":
		return serve(ctx)
	default:
		fmt.Printf("unknown command: %s\n", cmd)
		return nil
	}
}

func main() {
	if err := Main(context.Background()); err != nil {
		log.Fatal(err)
	}
}
