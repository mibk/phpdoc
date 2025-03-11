// Package phptype declares PHP types that can be used in PHPDoc
// syntax trees.
package phptype

// A Type is the interface that represents all PHP types.
type Type interface{ aType() }

type typ struct{}

func (*typ) aType() {}

// A Union represents a union of types.
type Union struct {
	typ
	Types []Type
}

// An Intersect represents an intersect of types.
type Intersect struct {
	typ
	Types []Type
}

// A Paren represents a parenthesized type.
type Paren struct {
	typ
	Type Type
}

// An Array represent an array of a specified type.
type Array struct {
	typ
	Elem Type
}

// Nullable represents a nullable type.
type Nullable struct {
	typ
	Type Type
}

// An ArrayShape represents the structure of key-values of a PHP array
// in the ordered-map mode.
type ArrayShape struct {
	typ
	Elems []*ArrayElem
}

// An ArrayElem represents a key-value element of ArrayShape.
type ArrayElem struct {
	Key      string
	Type     Type
	Optional bool
}

// An ObjectShape represents the structure of \stdClass.
type ObjectShape struct {
	typ
	Elems []*ObjectElem
}

// An ObjectElem represents a key-value element of ObjectShape.
type ObjectElem struct {
	Key      string
	Type     Type
	Optional bool
}

// Generic represents a pseudo-generic PHP type.
type Generic struct {
	typ
	Base       Type
	TypeParams []Type
}

type ConstFetch struct {
	typ
	Class Type
	Name  string
}

type Literal struct {
	typ
	Value string
}

// Named represents a (possibly qualified or fully qualified) PHP
// name, which might be a class name, a built-in type, or a special
// type (e.g. null, true).
type Named struct {
	typ
	Parts  []string
	Global bool // fully qualified
}

// This represents the $this special type.
type This struct{ typ }

type Param struct {
	Type     Type
	ByRef    bool // pass by reference
	Variadic bool
	Name     string
	Default  *Literal // or nil
}

type Callable struct {
	typ
	Params []*Param
	Result Type
}
