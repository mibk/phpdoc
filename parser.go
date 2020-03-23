package phpdoc

import (
	"fmt"
	"strings"
)

type PHPDoc struct {
	Lines []Line
}

func (doc *PHPDoc) String() string {
	var b strings.Builder
	b.WriteString("/**\n")
	for _, line := range doc.Lines {
		s := line.String()
		if s == "" {
			b.WriteString(" *\n")
		} else {
			b.WriteString(" * " + line.String() + "\n")
		}
	}
	b.WriteString(" */\n")
	return b.String()
}

type Line interface{ fmt.Stringer }

type TextLine struct {
	Value string
}

func (txt *TextLine) String() string {
	return txt.Value
}

type TagLine interface{ fmt.Stringer }

type ParamTag struct {
	Type PHPType
	Var  Token
	Desc string
}

func (tag *ParamTag) String() string {
	s := "@param " + tag.Type.String() + " " + tag.Var.Text
	if tag.Desc != "" {
		return s + " " + tag.Desc
	}
	return s
}

type ReturnTag struct {
	Type PHPType
	Desc string
}

func (tag *ReturnTag) String() string {
	s := "@return " + tag.Type.String()
	if tag.Desc != "" {
		return s + " " + tag.Desc
	}
	return s
}

type OtherTag struct {
	Name string
	Desc string
}

func (tag *OtherTag) String() string {
	s := "@" + tag.Name
	if tag.Desc != "" {
		return s + " " + tag.Desc
	}
	return s
}

type PHPType interface{ fmt.Stringer }

type PHPScalarType struct {
	Name string
}

func (t *PHPScalarType) String() string {
	return t.Name
}

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

REPEAT:
	for _, tt := range ttypes {
		if p.tok.Type == tt {
			p.next()
			goto REPEAT
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
	// TODO: ignore leading whitespace

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
	p.consume(Whitespace, Asterisk)
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
	tag.Var = p.tok
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
	name := p.tok.Text
	p.expect(Ident)
	return &PHPScalarType{Name: name}
}

func (p *Parser) parseDesc() string {
	var b strings.Builder
	for ; p.tok.Type != Newline && p.tok.Type != CloseDoc; p.next() {
		b.WriteString(p.tok.Text)
	}
	return strings.TrimSpace(b.String())
}
