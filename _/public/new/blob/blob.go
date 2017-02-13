package blob

import (
	"errors"
	"io"
	"sync"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jsbuiltin"
)

type Blob struct {
	mu     *sync.RWMutex
	blob   *js.Object
	offset int64
	size   int64
	mime   string
}

func New(o *js.Object) (*Blob, error) {
	if !jsbuiltin.InstanceOf(o, js.Global.Get("Blob")) {
		return nil, errors.New("object is not an instance of Blob")
	}
	return &Blob{
		mu:   &sync.RWMutex{},
		blob: o,
		size: o.Get("size").Int64(),
		mime: o.Get("type").String(),
	}, nil
}

func (b *Blob) Read(p []byte) (n int, err error) {
	b.mu.Lock()
	n, err = b.ReadAt(p, b.offset)
	b.offset += int64(n)
	b.mu.Unlock()
	return
}

func (b *Blob) ReadAt(p []byte, offset int64) (n int, err error) {
	if b.offset >= b.size {
		err = io.EOF
		return
	}
	length := offset + int64(len(p))
	if length > b.size {
		length = b.size
	}
	r := js.Global.Get("FileReader").New()
	ch := make(chan error)
	r.Set("onload", func(e *js.Object) {
		go func() { ch <- nil }()
	})
	r.Set("onerror", func(e *js.Object) {
		go func() { ch <- errors.New("could not load chunk") }()
	})
	r.Call("readAsArrayBuffer", b.blob.Call("slice", offset, length))
	err = <-ch
	bp := js.Global.Get("Uint8Array").New(r.Get("result")).Interface().([]byte)
	copy(p[:len(bp)], bp)
	n = len(bp)
	return
}

func (b *Blob) Size() int64 {
	return b.size
}

func (b *Blob) Mime() string {
	return b.mime
}

func (b *Blob) URL() URL {
	return URL(js.Global.Get("URL").Call("createObjectURL", b.blob).String())
}

func (b *Blob) Underlying() *js.Object {
	return b.blob
}

type URL string

func (u URL) Revoke() {
	js.Global.Get("URL").Call("revokeObjectURL", u)
}
