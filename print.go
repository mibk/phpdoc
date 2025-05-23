package phpdoc

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"mibk.dev/phpdoc/internal/token"
	"mibk.dev/phpdoc/phptype"
)

// Fprint "pretty-prints" an AST node to w.
func Fprint(w io.Writer, node interface{}) error {
	tw := tabwriter.NewWriter(w, 0, 0, 1, ' ', tabwriter.StripEscape)
	buf := bufio.NewWriter(tw)
	p := &printer{buf: buf}
	p.print(node)
	if p.err != nil {
		return p.err
	}
	if err := p.buf.Flush(); err != nil {
		return err
	}
	return tw.Flush()
}

type printer struct {
	buf *bufio.Writer
	err error // sticky
}

type whitespace byte

const (
	nextcol whitespace = '\v'
	tabesc  whitespace = tabwriter.Escape
	newline whitespace = '\n'
)

func (p *printer) print(args ...interface{}) {
	for _, arg := range args {
		if p.err != nil {
			return
		}

		switch arg := arg.(type) {
		case *Block:
			p.print(tabesc, arg.Indent, tabesc, token.OpenDoc)
			if arg.PreferOneline && len(arg.Lines) == 1 {
				p.print(arg.Lines[0])
			} else {
				p.print(newline)
				for _, line := range arg.Lines {
					p.print(tabesc, arg.Indent, tabesc, " *", line, newline)
				}
				p.print(tabesc, arg.Indent, tabesc)
			}
			p.print(' ', token.CloseDoc, newline)
		case Line:
			p.printLine(arg)
		case phptype.Type:
			p.printPHPType(arg)
		case []*phptype.Param:
			p.print(token.Lparen)
			for i, par := range arg {
				if i > 0 {
					p.print(token.Comma, ' ')
				}
				p.print(par.Type)
				if par.Name != "" {
					p.print(' ', par)
				}
			}
			p.print(token.Rparen)
		case *phptype.Param:
			// The type is printed by the owner.
			if arg.ByRef {
				p.print(token.And)
			}
			if arg.Variadic {
				p.print(token.Ellipsis)
			}
			p.print('$', arg.Name)
			if arg.Default != nil {
				p.print(' ', token.Assign, ' ', arg.Default)
			}
		case token.Type:
			_, p.err = p.buf.WriteString(arg.String())
		case string:
			_, p.err = p.buf.WriteString(arg)
		case rune:
			_, p.err = p.buf.WriteRune(arg)
		case whitespace:
			p.err = p.buf.WriteByte(byte(arg))
		default:
			p.err = fmt.Errorf("unsupported type %T", arg)
		}
	}
}

func (p *printer) printLine(line Line) {
	switch l := line.(type) {
	case *TextLine:
		if l.Value == "*" {
			return
		}
		if line := strings.TrimPrefix(l.Value, "* "); line != "" {
			p.print(tabesc, ' ', line, tabesc)
		}
	case Tag:
		p.print(' ')
		p.printTag(l)
	default:
		panic(fmt.Sprintf("unknown line type %T", line))
	}
}

func (p *printer) printTag(tag Tag) {
	switch tag := tag.(type) {
	case *ParamTag:
		p.print("@param", nextcol, tag.Param.Type, nextcol, tag.Param)
	case *ReturnTag:
		p.print("@return", nextcol, tag.Type)
	case *PropertyTag:
		p.print("@property")
		switch {
		case tag.ReadOnly && tag.WriteOnly:
			// Impossible, but…
		case tag.ReadOnly:
			p.print("-read")
		case tag.WriteOnly:
			p.print("-write")
		}
		p.print(nextcol, tag.Type, nextcol, '$', tag.Var)
	case *MethodTag:
		p.print("@method", nextcol)
		if tag.Static {
			p.print(token.Static, ' ')
		}
		if tag.Result != nil {
			p.print(tag.Result, ' ')
		}
		p.print(tag.Name, tag.Params)
	case *VarTag:
		p.print("@var", nextcol, tag.Type)
		if tag.Var != "" {
			p.print(nextcol, '$', tag.Var)
		}
	case *ThrowsTag:
		p.print("@throws", nextcol, tag.Class)
	case *ExtendsTag:
		p.print("@extends", nextcol, tag.Class)
	case *ImplementsTag:
		p.print("@implements", nextcol, tag.Interface)
	case *UsesTag:
		p.print("@uses", nextcol, tag.Trait)
	case *TemplateTag:
		p.print("@template", nextcol, tag.Param)
		if tag.Bound != nil {
			p.print(" of ", tag.Bound)
		}
	case *TypeDefTag:
		p.print("@phpstan-type", nextcol, tag.Name, nextcol, tag.Type)
	case *OtherTag:
		p.print('@', tag.Name)
	default:
		panic(fmt.Sprintf("unknown tag line %T", tag))
	}
	if desc := tag.desc(); desc != "" {
		p.print(nextcol, tabesc, desc, tabesc)
	}
}

func (p *printer) printPHPType(typ phptype.Type) {
	switch typ := typ.(type) {
	case *phptype.Union:
		for i, typ := range typ.Types {
			if i > 0 {
				p.print(token.Or)
			}
			p.print(typ)
		}
	case *phptype.Intersect:
		for i, typ := range typ.Types {
			if i > 0 {
				p.print(token.And)
			}
			p.print(typ)
		}
	case *phptype.Paren:
		p.print(token.Lparen, typ.Type, token.Rparen)
	case *phptype.Array:
		p.print(typ.Elem, token.Lbrack, token.Rbrack)
	case *phptype.Nullable:
		p.print(token.Qmark, typ.Type)
	case *phptype.Callable:
		p.print(token.Callable)
		if len(typ.Params) > 0 || typ.Result != nil {
			p.print(typ.Params)
			if typ.Result != nil {
				p.print(token.Colon, ' ', typ.Result)
			}
		}
	case *phptype.ArrayShape:
		p.print(token.Array)
		if len(typ.Elems) == 0 {
			break
		}
		p.print(token.Lbrace)
		for i, elem := range typ.Elems {
			if i > 0 {
				p.print(token.Comma, ' ')
			}
			if elem.Key != "" {
				p.print(elem.Key)
				if elem.Optional {
					p.print(token.Qmark)
				}
				p.print(token.Colon, ' ')
			}
			p.print(elem.Type)
		}
		p.print(token.Rbrace)
	case *phptype.ObjectShape:
		p.print(token.Object)
		if len(typ.Elems) == 0 {
			break
		}
		p.print(token.Lbrace)
		for i, elem := range typ.Elems {
			if i > 0 {
				p.print(token.Comma, ' ')
			}
			p.print(elem.Key)
			if elem.Optional {
				p.print(token.Qmark)
			}
			p.print(token.Colon, ' ', elem.Type)
		}
		p.print(token.Rbrace)
	case *phptype.Generic:
		p.print(typ.Base, token.Lt)
		for i, typ := range typ.TypeParams {
			if i > 0 {
				p.print(token.Comma, ' ')
			}
			p.print(typ)
		}
		p.print(token.Gt)
	case *phptype.ConstFetch:
		p.print(typ.Class, token.DoubleColon, typ.Name)
	case *phptype.Literal:
		p.print(typ.Value)
	case *phptype.Named:
		for i, part := range typ.Parts {
			if i > 0 || typ.Global {
				p.print(token.Backslash)
			}
			p.print(part)
		}
	case *phptype.This:
		p.print(token.This)
	default:
		panic(fmt.Sprintf("unknown PHP type %T", typ))
	}
}
