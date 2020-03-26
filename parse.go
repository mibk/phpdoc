package phpdoc

import (
	"fmt"
	"strings"
)

type Parser struct {
	sc *Scanner

	err error
	tok Token
}

func NewParser(sc *Scanner) *Parser {
	return &Parser{sc: sc}
}

func (p *Parser) Parse() (*PHPDoc, error) {
	doc := p.parseDoc()
	return doc, p.err
}

func (p *Parser) next() {
	p.tok = p.sc.Next()
}

func (p *Parser) expect(tt TokenType) {
	if p.tok.Type != tt {
		p.errorf("expecting %v, found %v", tt, p.tok)
	}
	p.next()
}

func (p *Parser) consume(ttypes ...TokenType) {
	if len(ttypes) == 0 {
		panic("not token types to consume provided")
	}

	for ; len(ttypes) > 0; ttypes = ttypes[1:] {
		if p.tok.Type == ttypes[0] {
			p.next()
		}
	}
}

func (p *Parser) errorf(format string, args ...interface{}) {
	if p.err == nil {
		p.err = fmt.Errorf(format, args...)
	}
}

// PHPDoc
//	'/**' [ newline ] Line [ newline Line ] ... '*/'
func (p *Parser) parseDoc() *PHPDoc {
	p.consume(Whitespace)
	p.next()
	p.expect(OpenDoc)
	lines := p.parseLines()
	p.expect(CloseDoc)
	return &PHPDoc{Lines: lines}
}

func (p *Parser) parseLines() []Line {
	var lines []Line
	p.consume(Newline)
	for {
		if p.err != nil {
			return nil
		}
		if p.tok.Type == CloseDoc {
			return lines
		}

		line := p.parseLine()
		lines = append(lines, line)

		switch p.tok.Type {
		case Newline:
			p.next()
			continue
		default:
			return lines
		}
	}
}

// Line
//	TextLine
//	TagLine
func (p *Parser) parseLine() Line {
	p.consume(Whitespace, Asterisk, Whitespace)
	if p.tok.Type == Tag {
		return p.parseTag()
	} else {
		return &TextLine{Value: p.parseDesc()}
	}
}

func (p *Parser) parseTag() TagLine {
	name := p.tok.Text
	p.expect(Tag)

	switch name {
	case "@param":
		return p.parseParamTag()
	case "@return":
		return p.parseReturnTag()
	default:
		return p.parseOtherTag(name[1:])
		return nil
	}
}

func (p *Parser) parseParamTag() *ParamTag {
	tag := new(ParamTag)
	p.consume(Whitespace)
	tag.Type = p.parseType()
	p.consume(Whitespace)
	if p.tok.Type == Ellipsis {
		tag.Variadic = true
		p.next()
		p.consume(Whitespace)
	}
	tag.Var = p.tok.Text[1:]
	p.expect(Var)
	tag.Desc = p.parseDesc()
	return tag
}

func (p *Parser) parseReturnTag() *ReturnTag {
	tag := new(ReturnTag)
	p.consume(Whitespace)
	tag.Type = p.parseType()
	tag.Desc = p.parseDesc()
	return tag
}

func (p *Parser) parseOtherTag(name string) *OtherTag {
	tag := &OtherTag{Name: name}
	tag.Desc = p.parseDesc()
	return tag
}

func (p *Parser) parseType() PHPType {
	typ := p.parseAtomicType()
	switch p.tok.Type {
	case Union:
		return p.parseUnionType(typ)
	case Intersect:
		return p.parseIntersectType(typ)
	}
	return typ
}

func (p *Parser) parseUnionType(init PHPType) PHPType {
	ut := &PHPUnionType{Types: make([]PHPType, 0, 2)}
	ut.Types = append(ut.Types, init)

	for p.tok.Type == Union {
		p.next()
		typ := p.parseAtomicType()
		ut.Types = append(ut.Types, typ)
	}
	return ut
}

func (p *Parser) parseIntersectType(init PHPType) PHPType {
	ut := &PHPIntersectType{Types: make([]PHPType, 0, 2)}
	ut.Types = append(ut.Types, init)

	for p.tok.Type == Intersect {
		p.next()
		typ := p.parseAtomicType()
		ut.Types = append(ut.Types, typ)
	}
	return ut
}

func (p *Parser) parseAtomicType() PHPType {
	typ := p.parseIdentType()
	if p.tok.Type == OpenAngle {
		p.next()
		typ = p.parseGenericType(typ)
	}
	if p.tok.Type == OpenBrack {
		p.next()
		p.expect(CloseBrack)
		return &PHPArrayType{Elem: typ}
	}
	return typ
}

func (p *Parser) parseGenericType(base PHPType) PHPType {
	var generics []PHPType
	for {
		t := p.parseType()
		generics = append(generics, t)
		p.consume(Whitespace)
		if p.tok.Type != Comma {
			break
		}
		p.next()
		p.consume(Whitespace)
	}
	p.expect(CloseAngle)
	return &PHPGenericType{Base: base, Generics: generics}
}

func (p *Parser) parseIdentType() PHPType {
	typ := new(PHPIdentType)
	if p.tok.Type == Nullable {
		typ.Nullable = true
		p.next()
		p.consume(Whitespace)
	}
	typ.Name = p.tok.Text
	p.expect(Ident)
	return typ
}

func (p *Parser) parseDesc() string {
	var b strings.Builder
	for ; p.tok.Type != Newline && p.tok.Type != CloseDoc; p.next() {
		b.WriteString(p.tok.Text)
	}
	return strings.TrimSpace(b.String())
}
