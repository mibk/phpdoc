package phpdoc

import (
	"fmt"
)

type PHPDoc struct {
	Lines []Line
}

type Line interface{ fmt.Stringer }

type TextLine struct {
	Value string
}

type TagLine interface{ fmt.Stringer }

type ParamTag struct {
	Type     PHPType
	Variadic bool
	Var      string
	Desc     string
}

type ReturnTag struct {
	Type PHPType
	Desc string
}

type PropertyTag struct {
	ReadOnly, WriteOnly bool
	Type                PHPType
	Desc                string
}

type OtherTag struct {
	Name string
	Desc string
}

type PHPType interface{ fmt.Stringer }

type PHPUnionType struct {
	Types []PHPType
}

type PHPIntersectType struct {
	Types []PHPType
}

type PHPParenType struct {
	Type PHPType
}

type PHPArrayType struct {
	Elem PHPType
}

type PHPGenericType struct {
	Base     PHPType
	Generics []PHPType
}

type PHPIdentType struct {
	Name     *PHPIdent
	Nullable bool
}

type PHPIdent struct {
	Parts  []string
	Global bool
}
