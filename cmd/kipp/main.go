package main

import (
	"context"
	"flag"
	"fmt"
	"log"
)

// Main makes writing programs easier by taking a context, and returning an
// error. It gives a more natural way to write mains.
func Main(ctx context.Context) error {
	switch cmd := flag.Arg(0); cmd {
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
