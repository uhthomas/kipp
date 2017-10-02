package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/dustin/go-humanize"
)

type UploadCommand struct {
	File    *os.File
	Private bool
	URL     *url.URL
}

func (u *UploadCommand) Do() {
	defer u.File.Close()
	u.URL.Path = "/upload"

	fi, err := u.File.Info()
	if err != nil {
		log.Fatal(err)
	}

	var pathPrefix string
	var r io.Reader = u.File
	if u.Private {
		key := make([]byte, 16)
		if _, err := io.ReadFull(rand.Reader, key); err != nil {
			log.Fatal(err)
		}
		c, err := aes.NewCipher(key)
		if err != nil {
			log.Fatal(err)
		}
		gcm, err := cipher.NewGCM(c)
		if err != nil {
			log.Fatal(err)
		}
		nonce := make([]byte, gcm.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			log.Fatal(err)
		}
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(u.File); err != nil {
			log.Fatal(err)
		}
		r = bytes.NewReader(gcm.Seal(nil, nonce, buf.Bytes(), nil))
		pathPrefix = "private#" + base64.RawURLEncoding.EncodeToString(append(nonce, key...)) + "/"
	}

	pr, pw := io.Pipe()
	w := multipart.NewWriter(pw)
	go func() {
		defer pw.Close()
		fw, err := w.CreateFormFile("file", fi.Name())
		if err != nil {
			pr.CloseWithError(err)
			return
		}
		if _, err := io.Copy(fw, r); err != nil {
			pr.CloseWithError(err)
			return
		}
		if err := w.Close(); err != nil {
			pr.CloseWithError(err)
			return
		}
	}()

	req, err := http.NewRequest(http.MethodPost, u.URL.String(), pr)
	if err != nil {
		log.Fatal(err)
	}
	// TODO: add content-length - nginx would 400 and it's hard to determine
	//       mime header size
	req.Header.Set("Content-Type", w.FormDataContentType())

	res, err := (&http.Client{}).Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	var out struct {
		Expires *time.Time `json:"expires,omitempty"`
		Path    string     `json:"path"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		log.Fatal(err)
	}

	u.URL.Path = ""
	fmt.Print(u.URL.String() + "/" + pathPrefix + out.Path)
	if out.Expires != nil {
		// print expiration and round time to compensate for server time and
		// request duration
		fmt.Printf(" expires %s\n", humanize.Time((*out.Expires).Round(time.Hour)))
	} else {
		fmt.Println(" uploaded permanently\n")
	}
}
