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

type PHPUnionType struct {
	Types []PHPType
}

func (t *PHPUnionType) String() string {
	var b strings.Builder
	for i, typ := range t.Types {
		if i > 0 {
			b.WriteRune('|')
		}
		b.WriteString(typ.String())
	}
	return b.String()
}

type PHPIdentType struct {
	Name string
}

func (t *PHPIdentType) String() string {
	return t.Name
}
