package phpdoc

import (
	"fmt"
	"io"
	"strings"

	"mibk.io/phpdoc/internal/token"
	"mibk.io/phpdoc/phptype"
)

// SyntaxError records an error and the position it occured on.
type SyntaxError struct {
	Line, Column int
	Err          error
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("line:%d:%d: %v", e.Line, e.Column, e.Err)
}

type parser struct {
	sc *token.Scanner

	err error
	tok token.Token
}

// Parse parses a single PHPDoc comment.
func Parse(r io.Reader) (*PHPDoc, error) {
	p := &parser{sc: token.NewScanner(r)}
	doc := p.parseDoc()
	if p.err != nil {
		return nil, p.err
	}
	return doc, nil
}

func (p *parser) next0() {
	p.tok = p.sc.Next()
}

// next is like next0 but skips whitespace.
func (p *parser) next() {
	p.next0()
	p.consume(token.Whitespace)
}

func (p *parser) expect(typ token.Type) {
	if p.tok.Type != typ {
		p.errorf("expecting %v, found %v", typ, p.tok)
	}
	p.next()
}

func (p *parser) got(typ token.Type) bool {
	if p.tok.Type == typ {
		p.next()
		return true
	}
	return false
}

func (p *parser) consume(types ...token.Type) {
	if len(types) == 0 {
		panic("not token types to consume provided")
	}

	for ; len(types) > 0; types = types[1:] {
		if p.tok.Type == types[0] {
			p.next0()
		}
	}
}

func (p *parser) errorf(format string, args ...interface{}) {
	if p.err == nil {
		se := &SyntaxError{Err: fmt.Errorf(format, args...)}
		se.Line, se.Column = p.tok.Pos.Line, p.tok.Pos.Column
		p.err = se
	}
}

// The syntax comments roughly follow the notation as defined at
// https://golang.org/ref/spec#Notation.

// PHPDoc = "/**" [ newline ] Line { newline Line } [ newline ] "*/" .
func (p *parser) parseDoc() *PHPDoc {
	doc := new(PHPDoc)
	p.next0()
	for {
		p.consume(token.Newline)
		if p.tok.Type != token.Whitespace {
			break
		}
		doc.Indent = p.tok.Text
		p.next0()
	}
	p.expect(token.OpenDoc)
	if !p.got(token.Newline) {
		doc.PreferOneline = true
	}
	doc.Lines = p.parseLines()
	p.expect(token.CloseDoc)
	return doc
}

func (p *parser) parseLines() []Line {
	var lines []Line
	for {
		if p.err != nil {
			return nil
		}
		if p.tok.Type == token.CloseDoc {
			return lines
		}

		line := p.parseLine()
		lines = append(lines, line)

		if !p.got(token.Newline) {
			return lines
		}
	}
}

// Line     = [ asterisk ] ( TextLine | Tag ) .
// TextLine = Desc .
func (p *parser) parseLine() Line {
	p.consume(token.Whitespace, token.Asterisk, token.Whitespace)
	if p.tok.Type == token.TagName {
		return p.parseTag()
	} else {
		return &TextLine{Value: p.parseDesc()}
	}
}

// Tag = ParamTag |
//       ReturnTag |
//       PropertyTag |
//       VarTag |
//       ThrowsTag |
//       ImplementsTag |
//       TemplateTag |
//       OtherTag .
func (p *parser) parseTag() Tag {
	name := p.tok.Text
	p.expect(token.TagName)

	switch name {
	case "@param":
		return p.parseParamTag()
	case "@return":
		return p.parseReturnTag()
	case "@property", "@property-read", "@property-write":
		return p.parsePropertyTag(name)
	case "@var":
		return p.parseVarTag()
	case "@throws":
		return p.parseThrowsTag()
	case "@implements":
		return p.parseImplementsTag()
	case "@template":
		return p.parseTemplateTag()
	default:
		return p.parseOtherTag(name[1:])
		return nil
	}
}

// ParamTag = "@param" PHPType [ "..." ] varname [ Desc ] .
func (p *parser) parseParamTag() *ParamTag {
	tag := new(ParamTag)
	tag.Type = p.parseType()
	if p.got(token.Ellipsis) {
		tag.Variadic = true
	}
	tag.Var = strings.TrimPrefix(p.tok.Text, "$")
	p.expect(token.VarName)
	tag.Desc = p.parseDesc()
	return tag
}

// ReturnTag = "@return" PHPType [ Desc ] .
func (p *parser) parseReturnTag() *ReturnTag {
	tag := new(ReturnTag)
	tag.Type = p.parseType()
	tag.Desc = p.parseDesc()
	return tag
}

// PropertyTag = ( "@property" | "@property-read" | "@property-write" ) PHPType varname [ Desc ] .
func (p *parser) parsePropertyTag(name string) *PropertyTag {
	tag := new(PropertyTag)
	tag.Type = p.parseType()
	tag.Var = strings.TrimPrefix(p.tok.Text, "$")
	p.expect(token.VarName)
	tag.Desc = p.parseDesc()

	switch {
	case strings.HasSuffix(name, "-read"):
		tag.ReadOnly = true
	case strings.HasSuffix(name, "-write"):
		tag.WriteOnly = true
	}
	return tag
}

// VarTag = "@var" PHPType [ varname ] [ Desc ] .
func (p *parser) parseVarTag() *VarTag {
	tag := new(VarTag)
	tag.Type = p.parseType()
	if p.tok.Type == token.VarName {
		tag.Var = p.tok.Text[1:]
		p.next()
	}
	tag.Desc = p.parseDesc()
	return tag
}

// ThrowsTag = "@throws" PHPType [ Desc ] .
func (p *parser) parseThrowsTag() *ThrowsTag {
	tag := new(ThrowsTag)
	tag.Class = p.parseType()
	tag.Desc = p.parseDesc()
	return tag
}

// ImplementsTag = "@implements" PHPType [ Desc ] .
func (p *parser) parseImplementsTag() *ImplementsTag {
	tag := new(ImplementsTag)
	tag.Interface = p.parseType()
	tag.Desc = p.parseDesc()
	return tag
}

// TemplateTag = "@template" ident [ ( "of | "as" ) PHPType ] [ Desc ] .
func (p *parser) parseTemplateTag() *TemplateTag {
	tag := new(TemplateTag)
	tag.Param = p.tok.Text
	p.expect(token.Ident)
	if p.tok.Type == token.Ident && (p.tok.Text == "of" || p.tok.Text == "as") {
		p.next()
		tag.Bound = p.parseType()
	}
	tag.Desc = p.parseDesc()
	return tag
}

// OtherTag = tagname [ Desc ] .
func (p *parser) parseOtherTag(name string) *OtherTag {
	tag := &OtherTag{Name: name}
	tag.Desc = p.parseDesc()
	return tag
}

// PHPType = AtomicType | UnionType | IntersectType .
func (p *parser) parseType() phptype.Type {
	typ := p.parseAtomicType()
	switch p.tok.Type {
	case token.Or:
		return p.parseUnionType(typ)
	case token.And:
		return p.parseIntersectType(typ)
	}
	return typ
}

// UnionType = AtomicType "|" AtomicType { "|" AtomicType } .
func (p *parser) parseUnionType(init phptype.Type) phptype.Type {
	union := &phptype.Union{Types: make([]phptype.Type, 0, 2)}
	union.Types = append(union.Types, init)

	for p.got(token.Or) {
		typ := p.parseAtomicType()
		union.Types = append(union.Types, typ)
	}
	return union
}

// IntersectType = AtomicType "&" AtomicType { "&" AtomicType } .
func (p *parser) parseIntersectType(init phptype.Type) phptype.Type {
	intersect := &phptype.Intersect{Types: make([]phptype.Type, 0, 2)}
	intersect.Types = append(intersect.Types, init)

	for p.got(token.And) {
		typ := p.parseAtomicType()
		intersect.Types = append(intersect.Types, typ)
	}
	return intersect
}

// AtomicType   = ParenType | NullableType | ArrayType .
// NullableType = [ "?" ] ( GenericType | BasicType ) .
// BasicType    = IdentType | ArrayShapeType .
// ArrayType    = AtomicType "[" "]" .
func (p *parser) parseAtomicType() phptype.Type {
	var typ phptype.Type
	if p.got(token.Lparen) {
		typ = p.parseParenType()
	} else {
		nullable := p.got(token.Qmark)
		if p.got(token.Array) {
			typ = p.parseArrayShapeType()
		} else {
			typ = p.parseIdentType()
		}
		// TODO: Forbid generic params for arrays with a shape?
		if p.got(token.Lt) {
			typ = p.parseGenericType(typ)
		}
		if nullable {
			typ = &phptype.Nullable{Type: typ}
		}
	}
	for p.got(token.Lbrack) {
		p.expect(token.Rbrack)
		typ = &phptype.Array{Elem: typ}
	}
	return typ
}

// ParenType = "(" PHPType ")" .
func (p *parser) parseParenType() phptype.Type {
	typ := new(phptype.Paren)
	typ.Type = p.parseType()
	p.expect(token.Rparen)
	return typ
}

// ArrayShapeType = array [ ArrayShape ] .
// ArrayShape     = "{" KeyType { "," KeyType } "}" .
// KeyType        = ArrayKey [ "?" ] ":" PHPType .
// ArrayKey       = ident | decimal .
func (p *parser) parseArrayShapeType() phptype.Type {
	typ := new(phptype.ArrayShape)
	if p.got(token.Lbrace) {
		for {
			elem := new(phptype.ArrayElem)
			switch p.tok.Type {
			case token.Ident, token.Decimal:
				elem.Key = p.tok.Text
				p.next()
			default:
				// TODO: Consider not requiring array keys.
				p.errorf("expecting %v or %v, found %v", token.Ident, token.Decimal, p.tok)
				return nil
			}
			elem.Optional = p.got(token.Qmark)
			p.expect(token.Colon)
			elem.Type = p.parseType()
			typ.Elems = append(typ.Elems, elem)
			if !p.got(token.Comma) {
				break
			}
		}
		p.expect(token.Rbrace)
	}
	return typ
}

// GenericType = BasicType "<" PHPType { "," PHPType } ">" .
func (p *parser) parseGenericType(base phptype.Type) phptype.Type {
	var params []phptype.Type
	for {
		t := p.parseType()
		params = append(params, t)
		if !p.got(token.Comma) {
			break
		}
	}
	p.expect(token.Gt)
	return &phptype.Generic{Base: base, TypeParams: params}
}

// IdentType = [ "\\" ] ident { "\\" ident } .
func (p *parser) parseIdentType() *phptype.Ident {
	id := new(phptype.Ident)
	if p.got(token.Backslash) {
		id.Global = true
	}
	for {
		id.Parts = append(id.Parts, p.tok.Text)
		p.expect(token.Ident)
		if !p.got(token.Backslash) {
			break
		}
	}
	return id
}

// Desc = { any } .
func (p *parser) parseDesc() string {
	var b strings.Builder
LOOP:
	for {
		switch p.tok.Type {
		case token.Newline, token.CloseDoc, token.EOF:
			break LOOP
		}
		b.WriteString(p.tok.Text)
		p.next0()
	}
	return strings.TrimSpace(b.String())
}
