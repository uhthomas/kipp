package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
)

type UploadCommand struct {
	File     *os.File
	Insecure bool
	Private  bool
	URL      *url.URL
}

func (u *UploadCommand) Do() {
	defer u.File.Close()

	s, err := u.File.Stat()
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
		pathPrefix = "private#" + base64.RawURLEncoding.EncodeToString(append(nonce, key...))
	}

	pr, pw := io.Pipe()
	w := multipart.NewWriter(pw)
	go func() {
		defer pw.Close()
		fw, err := w.CreateFormFile("file", s.Name())
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
	c := &http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	if u.Insecure {
		c.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	res, err := c.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	rurl, err := res.Location()
	if err != nil {
		log.Fatal(err)
	}
	p := rurl.Path
	rurl.Path = ""
	fmt.Println(rurl.String() + "/" + pathPrefix + p)
}
