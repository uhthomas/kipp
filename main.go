package main

import (
	"conf/route"
	_ "conf/worker"
	"encoding/json"
	"log"
	"mime"
	"os"
	"path/filepath"
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
	check(os.MkdirAll(filepath.Join("_", "files"), 0777))
	check(os.RemoveAll(filepath.Join("_", "tmp")))
	check(os.MkdirAll(filepath.Join("_", "tmp"), 0777))
	check(initmimes())
	check(route.Listen())
}
