package phpdoc

import "mibk.io/phpdoc/phptype"

type PHPDoc struct {
	Lines         []Line
	Indent        string // … each line
	PreferOneline bool
}

type Line interface{ aLine() }

type line struct{}

func (*line) aLine() {}

type TextLine struct {
	line
	Value string
}

type Tag interface {
	Line
	aTag()
	desc() string
}

type tag struct{ line }

func (*tag) aTag() {}

type ParamTag struct {
	tag
	Type     phptype.Type
	Variadic bool
	Var      string
	Desc     string
}

type ReturnTag struct {
	tag
	Type phptype.Type
	Desc string
}

type PropertyTag struct {
	tag
	ReadOnly, WriteOnly bool
	Type                phptype.Type
	Desc                string
}

type VarTag struct {
	tag
	Type phptype.Type
	Var  string
	Desc string
}

type TemplateTag struct {
	tag
	Param string
	Bound phptype.Type // or nil
	Desc  string
}

type OtherTag struct {
	tag
	Name string
	Desc string
}

func (t *ParamTag) desc() string    { return t.Desc }
func (t *ReturnTag) desc() string   { return t.Desc }
func (t *PropertyTag) desc() string { return t.Desc }
func (t *VarTag) desc() string      { return t.Desc }
func (t *TemplateTag) desc() string { return t.Desc }
func (t *OtherTag) desc() string    { return t.Desc }
