package phpdoc

import "mibk.io/phpdoc/phptype"

// A Block represents a PHPDoc comment block.
type Block struct {
	Lines         []Line
	Indent        string // … each line
	PreferOneline bool
}

// A Line represents a line in a PHPDoc comment.
type Line interface{ aLine() }

type line struct{}

func (*line) aLine() {}

// A TextLine represents a regular, text-only line in a PHPDoc comment.
type TextLine struct {
	line
	Value string
}

// A Tag represents a tag line in a PHPDoc comment (e.g. @author).
type Tag interface {
	Line
	aTag()
	desc() string
}

type tag struct{ line }

func (*tag) aTag() {}

// A ParamTag represents a @param tag.
type ParamTag struct {
	tag
	Type     phptype.Type
	ByRef    bool // pass by reference
	Variadic bool
	Var      string
	Desc     string
}

// A ReturnTag represents a @return tag.
type ReturnTag struct {
	tag
	Type phptype.Type
	Desc string
}

// A PropertyTag represents a @property tag, as well as its variants
// @property-read and @property-write.
type PropertyTag struct {
	tag
	ReadOnly, WriteOnly bool
	Type                phptype.Type
	Var                 string
	Desc                string
}

// A VarTag represents a @var tag.
type VarTag struct {
	tag
	Type phptype.Type
	Var  string
	Desc string
}

// A ThrowsTag represents a @throws tag.
type ThrowsTag struct {
	tag
	Class phptype.Type
	Desc  string
}

// An ImplementsTag represents a @implements tag.
type ImplementsTag struct {
	tag
	Interface phptype.Type
	Desc      string
}

// A TemplateTag represents a @template tag.
type TemplateTag struct {
	tag
	Param string
	Bound phptype.Type // or nil
	Desc  string
}

// A OtherTag represents an arbitrary tag without a special meaning.
type OtherTag struct {
	tag
	Name string
	Desc string
}

func (t *ParamTag) desc() string      { return t.Desc }
func (t *ReturnTag) desc() string     { return t.Desc }
func (t *PropertyTag) desc() string   { return t.Desc }
func (t *VarTag) desc() string        { return t.Desc }
func (t *ThrowsTag) desc() string     { return t.Desc }
func (t *ImplementsTag) desc() string { return t.Desc }
func (t *TemplateTag) desc() string   { return t.Desc }
func (t *OtherTag) desc() string      { return t.Desc }
