// +build gofuzz

package phpdoc_test

import (
	"bytes"
	"io/ioutil"

	"mibk.io/phpdoc"
)

func Fuzz(b []byte) int {
	doc, err := phpdoc.Parse(bytes.NewReader(b))
	if err != nil {
		if doc != nil {
			panic("doc != nil on error")
		}
		return 0
	}
	if err := phpdoc.Fprint(ioutil.Discard, doc); err != nil {
		panic("cannot print")
	}
	return 1
}
