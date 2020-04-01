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

func (p *Parser) nextTok() {
	p.tok = p.sc.Next()
}

func (p *Parser) next() {
	p.nextTok()
	p.consume(Whitespace)
}

func (p *Parser) expect(tt TokenType) {
	if p.tok.Type != tt {
		p.errorf("expecting %v, found %v", tt, p.tok)
	}
	p.next()
}

func (p *Parser) got(tt TokenType) bool {
	if p.tok.Type == tt {
		p.next()
		return true
	}
	return false
}

func (p *Parser) consume(ttypes ...TokenType) {
	if len(ttypes) == 0 {
		panic("not token types to consume provided")
	}

	for ; len(ttypes) > 0; ttypes = ttypes[1:] {
		if p.tok.Type == ttypes[0] {
			p.nextTok()
		}
	}
}

func (p *Parser) errorf(format string, args ...interface{}) {
	if p.err == nil {
		p.err = fmt.Errorf(format, args...)
	}
}

// The syntax comments roughly follow the notation as defined at
// https://golang.org/ref/spec#Notation.

// PHPDoc = "/**" [ newline ] Line { newline Line } [ newline ] "*/" .
func (p *Parser) parseDoc() *PHPDoc {
	doc := new(PHPDoc)
	p.nextTok()
	for {
		p.consume(Newline)
		if p.tok.Type != Whitespace {
			break
		}
		doc.Indent = p.tok.Text
		p.nextTok()
	}
	p.expect(OpenDoc)
	if !p.got(Newline) {
		doc.PreferOneline = true
	}
	doc.Lines = p.parseLines()
	p.expect(CloseDoc)
	return doc
}

func (p *Parser) parseLines() []Line {
	var lines []Line
	for {
		if p.err != nil {
			return nil
		}
		if p.tok.Type == CloseDoc {
			return lines
		}

		line := p.parseLine()
		lines = append(lines, line)

		if !p.got(Newline) {
			return lines
		}
	}
}

// Line     = TextLine | TagLine .
// TextLine = Desc .
func (p *Parser) parseLine() Line {
	p.consume(Whitespace, Asterisk, Whitespace)
	if p.tok.Type == Tag {
		return p.parseTag()
	} else {
		return &TextLine{Value: p.parseDesc()}
	}
}

// TagLine = ParamTag |
//           ReturnTag |
//           PropertyTag |
//           VarTag |
//           OtherTag .
func (p *Parser) parseTag() TagLine {
	name := p.tok.Text
	p.expect(Tag)

	switch name {
	case "@param":
		return p.parseParamTag()
	case "@return":
		return p.parseReturnTag()
	case "@property", "@property-read", "@property-write":
		return p.parsePropertyTag(name)
	case "@var":
		return p.parseVarTag()
	default:
		return p.parseOtherTag(name[1:])
		return nil
	}
}

// ParamTag = "@param" PHPType [ "..." ] var [ Desc ] .
func (p *Parser) parseParamTag() *ParamTag {
	tag := new(ParamTag)
	tag.Type = p.parseType()
	if p.got(Ellipsis) {
		tag.Variadic = true
	}
	tag.Var = p.tok.Text[1:]
	p.expect(Var)
	tag.Desc = p.parseDesc()
	return tag
}

// ReturnTag = "@return" PHPType [ Desc ] .
func (p *Parser) parseReturnTag() *ReturnTag {
	tag := new(ReturnTag)
	tag.Type = p.parseType()
	tag.Desc = p.parseDesc()
	return tag
}

// PropertyTag = ( "@property" | "@property-read" | "@property-write" ) PHPType var [ Desc ] .
func (p *Parser) parsePropertyTag(name string) *PropertyTag {
	tag := new(PropertyTag)
	tag.Type = p.parseType()
	tag.Desc = p.parseDesc()

	switch {
	case strings.HasSuffix(name, "-read"):
		tag.ReadOnly = true
	case strings.HasSuffix(name, "-write"):
		tag.WriteOnly = true
	}
	return tag
}

// VarTag = "@var" PHPType var [ Desc ] .
func (p *Parser) parseVarTag() *VarTag {
	tag := new(VarTag)
	tag.Type = p.parseType()
	tag.Var = p.tok.Text[1:]
	p.expect(Var)
	tag.Desc = p.parseDesc()
	return tag
}

// OtherTag = tag [ Desc ] .
func (p *Parser) parseOtherTag(name string) *OtherTag {
	tag := &OtherTag{Name: name}
	tag.Desc = p.parseDesc()
	return tag
}

// PHPType = AtomicType | UnionType | IntersectType .
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

// UnionType = AtomicType "|" AtomicType { "|" AtomicType } .
func (p *Parser) parseUnionType(init PHPType) PHPType {
	ut := &PHPUnionType{Types: make([]PHPType, 0, 2)}
	ut.Types = append(ut.Types, init)

	for p.got(Union) {
		typ := p.parseAtomicType()
		ut.Types = append(ut.Types, typ)
	}
	return ut
}

// IntersectType = AtomicType "&" AtomicType { "&" AtomicType } .
func (p *Parser) parseIntersectType(init PHPType) PHPType {
	ut := &PHPIntersectType{Types: make([]PHPType, 0, 2)}
	ut.Types = append(ut.Types, init)

	for p.got(Intersect) {
		typ := p.parseAtomicType()
		ut.Types = append(ut.Types, typ)
	}
	return ut
}

// AtomicType = ParenType | GenericType | IdentType | ArrayType .
// ArrayType  = AtomicType "[" "]" .
func (p *Parser) parseAtomicType() PHPType {
	var typ PHPType
	if p.got(OpenParen) {
		typ = p.parseParenType()
	} else {
		typ = p.parseIdentType()
		if p.got(OpenAngle) {
			typ = p.parseGenericType(typ)
		}
	}
	for p.got(OpenBrack) {
		p.expect(CloseBrack)
		typ = &PHPArrayType{Elem: typ}
	}
	return typ
}

// ParenType = "(" PHPType ")" .
func (p *Parser) parseParenType() PHPType {
	t := new(PHPParenType)
	t.Type = p.parseType()
	p.expect(CloseParen)
	return t
}

// GenericType = IdentType "<" PHPType { "," PHPType } ">" .
func (p *Parser) parseGenericType(base PHPType) PHPType {
	var generics []PHPType
	for {
		t := p.parseType()
		generics = append(generics, t)
		if !p.got(Comma) {
			break
		}
	}
	p.expect(CloseAngle)
	return &PHPGenericType{Base: base, Generics: generics}
}

// IdentType = [ "?" ] IdentName .
func (p *Parser) parseIdentType() PHPType {
	typ := new(PHPIdentType)
	if p.got(Nullable) {
		typ.Nullable = true
	}
	typ.Name = p.parseIdentName()
	return typ
}

// IdentName = [ "\\" ] ident { "\\" ident } .
func (p *Parser) parseIdentName() *PHPIdent {
	id := new(PHPIdent)
	if p.got(Backslash) {
		id.Global = true
	}
	for {
		id.Parts = append(id.Parts, p.tok.Text)
		p.expect(Ident)
		if !p.got(Backslash) {
			break
		}
	}
	return id
}

// Desc = { any } .
func (p *Parser) parseDesc() string {
	var b strings.Builder
	for ; p.tok.Type != Newline && p.tok.Type != CloseDoc; p.nextTok() {
		b.WriteString(p.tok.Text)
	}
	return strings.TrimSpace(b.String())
}
