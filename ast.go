package phpdoc

type PHPDoc struct {
	Lines []Line
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

type PHPGenericType struct {
	phpType
	Base     PHPType
	Generics []PHPType
}

type PHPIdentType struct {
	phpType
	Name     *PHPIdent
	Nullable bool
}

type PHPIdent struct {
	phpType
	Parts  []string
	Global bool
}
