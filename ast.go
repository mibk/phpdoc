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

type TagLine interface {
	Line
	aTag()
}

type tagLine struct{ line }

func (*tagLine) aTag() {}

type ParamTag struct {
	tagLine
	Type     PHPType
	Variadic bool
	Var      string
	Desc     string
}

type ReturnTag struct {
	tagLine
	Type PHPType
	Desc string
}

type PropertyTag struct {
	tagLine
	ReadOnly, WriteOnly bool
	Type                PHPType
	Desc                string
}

type VarTag struct {
	tagLine
	Type PHPType
	Var  string
	Desc string
}

type OtherTag struct {
	tagLine
	Name string
	Desc string
}

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
	Base     PHPType
	Generics []PHPType
}

type PHPIdentType struct {
	phpType
	Parts  []string
	Global bool
}
