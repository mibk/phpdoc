package phpdoc

import (
	"io"

	"mibk.dev/phpdoc/internal/token"
	"mibk.dev/phpdoc/phptype"
)

func ParseType(r io.Reader) (phptype.Type, error) {
	p := &parser{scan: token.NewScanner(r)}
	p.next()
	typ := p.parseType()
	return typ, p.err
}
