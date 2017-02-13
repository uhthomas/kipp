package main

import (
	"encoding/hex"
	"io"
	"io/ioutil"
	"log"
	"sync"
	"time"

	blake2b "github.com/minio/blake2b-simd"

	"./blob"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"honnef.co/go/js/dom"
)

type percentReader struct {
	io.Reader
	mu          *sync.Mutex
	total, size int
}

func newPercentReader(r io.Reader, size int) *percentReader {
	return &percentReader{
		Reader: r,
		mu:     &sync.Mutex{},
		size:   size,
	}
}

func (p *percentReader) Read(b []byte) (n int, err error) {
	p.mu.Lock()
	n, err = p.Reader.Read(b)
	p.total += n
	println((float64(p.total) / float64(p.size)) * 100.0)
	p.mu.Unlock()
	return
}

var jq = jquery.NewJQuery

func preventDefault(e dom.Event) {
	e.PreventDefault()
}

func upload(f interface{}) {

}

func request(title string) bool {
	c := make(chan bool)
	el := jq(".confirmation-modal")
	el.Find(".title").SetText(title)
	yel := el.Find(".button.yes").On("click", func(e jquery.Event) {
		el.Hide()
		go func() { c <- true }()
	})
	defer yel.Off("click")
	nel := el.Find(".button.no").On("click", func(e jquery.Event) {
		el.Hide()
		go func() { c <- false }()
	})
	defer nel.Off("click")
	el.Show()
	return <-c
}

func buildBlobs(o *js.Object) (bs []*blob.Blob, err error) {
	for i := 0; i < o.Length(); i++ {
		b, err := blob.New(o.Index(i))
		if err != nil {
			return nil, err
		}
		bs = append(bs, b)
	}
	return
}

func cp(w io.Writer, r io.Reader) (n int64, err error) {
	b := make([]byte, 0, 256<<10)
	for {
		t := time.Now()
		nr, err := r.Read(b[:cap(b)])
		n += int64(nr)
		switch {
		case err == io.EOF:
			return n, nil
		case err != nil:
			return n, err
		}
		if _, err := w.Write(b[:nr]); err != nil {
			return n, err
		}
		println(time.Since(t).String())
	}
}

func process(o *js.Object) {
	files, err := buildBlobs(o)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		f := f
		go func() {
			h := blake2b.New512()
			r := newPercentReader(f, f.Underlying().Get("size").Int())
			// io.Copy(h, r)
			cp(ioutil.Discard, r)
			println(hex.EncodeToString(h.Sum(nil)))
		}()
	}
	if len(files) == 0 {
		return
	}
	if len(files) == 1 {
		upload(files[0])
		return
	}

}

func main() {
	document := dom.GetWindow().Document()
	document.AddEventListener("dragover", false, preventDefault)
	document.AddEventListener("dragenter", false, preventDefault)
	document.AddEventListener("dragend", false, preventDefault)
	document.AddEventListener("dragleave", false, preventDefault)
	document.AddEventListener("drop", false, func(e dom.Event) {
		e.PreventDefault()
		go process(e.Underlying().Get("dataTransfer").Get("files"))
	})
	document.QuerySelector("input").
		AddEventListener("change", false, func(e dom.Event) {
			e.PreventDefault()
			go process(e.Underlying().Get("target").Get("files"))
		})
}
