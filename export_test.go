package phpdoc

import (
	"io"

	"mibk.io/phpdoc/internal/token"
	"mibk.io/phpdoc/phptype"
)

func ParseType(r io.Reader) (phptype.Type, error) {
	p := &parser{sc: token.NewScanner(r)}
	p.next()
	typ := p.parseType()
	return typ, p.err
}
