package main

import (
	"flag"

	"github.com/alecthomas/units"
)

type bytesValue units.Base2Bytes

func newBytesValue(val units.Base2Bytes, p *units.Base2Bytes) *bytesValue {
	*p = val
	return (*bytesValue)(p)
}

func flagBytesValue(name string, value units.Base2Bytes, usage string) *units.Base2Bytes {
	p := new(units.Base2Bytes)
	flag.CommandLine.Var(newBytesValue(value, p), name, usage)
	return p
}

func (d *bytesValue) Set(s string) (err error) {
	*(*units.Base2Bytes)(d), err = units.ParseBase2Bytes(s)
	return err
}

func (d *bytesValue) Get() interface{} { return units.Base2Bytes(*d) }

func (d *bytesValue) String() string { return (*units.Base2Bytes)(d).String() }
