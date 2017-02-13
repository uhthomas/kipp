package main

import (
	"encoding/json"
	"log"
	"mime"
	"os"
	"path/filepath"

	"github.com/6f7262/conf/route"
	_ "github.com/6f7262/conf/worker"
)

func initmimes() error {
	f, err := os.Open(filepath.Join("_", "mime.json"))
	if err != nil {
		return err
	}
	var mimes map[string][]string
	if err := json.NewDecoder(f).Decode(&mimes); err != nil {
		return err
	}
	for m, es := range mimes {
		for _, e := range es {
			mime.AddExtensionType(e, m)
		}
	}
	return nil
}

func check(err error) {
	if err == nil {
		return
	}
	log.Fatal(err)
}

func main() {
	check(os.RemoveAll("_/files/tmp"))
	check(os.MkdirAll("_/files/tmp", 0777))
	check(initmimes())
	check(route.Listen())
}
