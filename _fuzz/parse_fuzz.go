// +build gofuzz

package phpdoc_test

import (
	"bytes"

	"mibk.io/phpdoc"
)

func Fuzz(b []byte) int {
	p := phpdoc.NewParser(bytes.NewReader(b))
	doc, err := p.Parse()
	if err != nil {
		if doc != nil {
			panic("doc != nil on error")
		}
		return 0
	}
	return 1
}
