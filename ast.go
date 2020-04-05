package phpdoc

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
	Description() string
}

type tag struct{ line }

func (*tag) aTag() {}

type ParamTag struct {
	tag
	Type     PHPType
	Variadic bool
	Var      string
	Desc     string
}

type ReturnTag struct {
	tag
	Type PHPType
	Desc string
}

type PropertyTag struct {
	tag
	ReadOnly, WriteOnly bool
	Type                PHPType
	Desc                string
}

type VarTag struct {
	tag
	Type PHPType
	Var  string
	Desc string
}

type TemplateTag struct {
	tag
	Param string
	Bound PHPType // or nil
	Desc  string
}

type OtherTag struct {
	tag
	Name string
	Desc string
}

func (t *ParamTag) Description() string    { return t.Desc }
func (t *ReturnTag) Description() string   { return t.Desc }
func (t *PropertyTag) Description() string { return t.Desc }
func (t *VarTag) Description() string      { return t.Desc }
func (t *TemplateTag) Description() string { return t.Desc }
func (t *OtherTag) Description() string    { return t.Desc }

type PHPType interface{ aType() }

type phpType struct{}

func (*phpType) aType() {}

type PHPUnionType struct {
	phpType
	Types []PHPType
}

type PHPIntersectType struct {
	phpType
	Types []PHPType
}

type PHPParenType struct {
	phpType
	Type PHPType
}

type PHPArrayType struct {
	phpType
	Elem PHPType
}

type PHPNullableType struct {
	phpType
	Type PHPType
}

type PHPArrayShapeType struct {
	phpType
	Elems []*PHPArrayElem
}

type PHPArrayElem struct {
	Key      string
	Type     PHPType
	Optional bool
}

type PHPGenericType struct {
	phpType
	Base       PHPType
	TypeParams []PHPType
}

type PHPIdentType struct {
	phpType
	Parts  []string
	Global bool
}
