package phpdoc

import "mibk.io/phpdoc/phptype"

func (p *Parser) ParseType() (phptype.Type, error) {
	p.next()
	typ := p.parseType()
	return typ, p.err
}
