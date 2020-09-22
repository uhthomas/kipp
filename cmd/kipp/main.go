package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// Main makes writing programs easier by taking a context, and returning an
// error. It gives a more natural way to write mains.
func Main(ctx context.Context) error {
	var cmd string
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	if len(cmd) > 0 && cmd[0] == '-' {
		cmd = ""
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
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		select {
		case <-ctx.Done():
		case <-c:
		}
		cancel()
	}()

	if err := Main(ctx); err != nil {
		log.Fatal(err)
	}
}
