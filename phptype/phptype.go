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

// Generic represents a pseudo-generic PHP type.
type Generic struct {
	typ
	Base       Type
	TypeParams []Type
}

// An Ident represents a (possibly qualified or fully qualified) PHP
// identifier, which might be a class, built-in type, or a special
// keyword (e.g. null, true).
type Ident struct {
	typ
	Parts  []string
	Global bool // fully qualified
}

type Param struct {
	typ
	Type     Type
	ByRef    bool // pass by reference
	Variadic bool
	Var      string
}

type Callable struct {
	typ
	Params []*Param
	Result Type
}
