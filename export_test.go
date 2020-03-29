package phpdoc

func (p *Parser) ParseType() (PHPType, error) {
	p.next()
	typ := p.parseType()
	return typ, p.err
}
