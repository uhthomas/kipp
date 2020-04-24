package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/alecthomas/units"
	"github.com/uhthomas/kipp"
	"github.com/uhthomas/kipp/filesystem/local"
	"gopkg.in/alecthomas/kingpin.v2"
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

func loadMimeTypes(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	m := make(map[string][]string)
	if err := json.NewDecoder(f).Decode(&m); err != nil {
		return err
	}
	for k, v := range m {
		for _, vv := range v {
			mime.AddExtensionType(vv, k)
		}
	}
	return nil
}

func Main(ctx context.Context) error {
	var s kipp.Server
	servecmd := kingpin.Command("serve", "Start a kipp server.").Default()

	addr := servecmd.
		Flag("addr", "Server listen address.").
		Default(":80").
		String()
	// cleanupInterval := servecmd.
	// 	Flag("cleanup-interval", "Cleanup interval for deleting expired files.").
	// 	Default("5m").
	// 	Duration()
	mime := servecmd.
		Flag("mime", "A json formatted collection of extensions and mime types.").
		PlaceHolder("PATH").
		String()
	servecmd.
		Flag("lifetime", "File lifetime, 0=infinite.").
		Default("24h").
		DurationVar(&s.Lifetime)
	servecmd.
		Flag("limit", "The maximum file size for uploads.").
		Default("150MB").
		BytesVar((*units.Base2Bytes)(&s.Limit))
	path := servecmd.
		Flag("files", "File path.").
		Default("files").
		String()
	tempPath := servecmd.
		Flag("tmp", "Temp path for in-progress uploads.").
		Default("files/tmp").
		String()
	servecmd.
		Flag("public", "Public path for web resources.").
		Default("public").
		StringVar(&s.PublicPath)

	var u UploadCommand
	{
		uploadcmd := kingpin.Command("upload", "Upload a file.")
		uploadcmd.
			Arg("file", "File to be uploaded").
			Required().
			FileVar(&u.File)
		uploadcmd.
			Flag("insecure", "Don't verify SSL certificates.").
			BoolVar(&u.Insecure)
		uploadcmd.
			Flag("private", "Encrypt the uploaded file").
			BoolVar(&u.Private)
		uploadcmd.
			Flag("url", "Source URL").
			Envar("kipp-upload-url").
			Default("https://kipp.6f.io").
			URLVar(&u.URL)
	}

	t := kingpin.Parse()

	// kipp upload
	if t == "upload" {
		u.Do()
		return nil
	}

	fs, err := local.New(*path, *tempPath)
	if err != nil {
		return fmt.Errorf("new local filesystem: %w", err)
	}
	s.FileSystem = fs

	// Load mime types
	if m := *mime; m != "" {
		if err := loadMimeTypes(m); err != nil {
			return fmt.Errorf("load mime types: %w", err)
		}
	}

	// // Start cleanup worker
	// if s.Lifetime > 0 {
	// 	w := worker(*cleanupInterval)
	// 	go w.Do(context.Background(), s.Cleanup)
	// }

	log.Printf("Listening on %s", *addr)
	return (&http.Server{
		Addr:    *addr,
		Handler: s,
		// ReadTimeout:  5 * time.Second,
		// WriteTimeout: 10 * time.Second,
		IdleTimeout: 120 * time.Second,
		BaseContext: func(net.Listener) context.Context { return ctx },
	}).ListenAndServe()
}

func main() {
	if err := Main(context.Background()); err != nil {
		log.Fatal(err)
	}
}
