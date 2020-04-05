package phptype

type Type interface{ aType() }

type typ struct{}

func (*typ) aType() {}

type Union struct {
	typ
	Types []Type
}

type Intersect struct {
	typ
	Types []Type
}

type Paren struct {
	typ
	Type Type
}

type Array struct {
	typ
	Elem Type
}

type Nullable struct {
	typ
	Type Type
}

type ArrayShape struct {
	typ
	Elems []*ArrayElem
}

type ArrayElem struct {
	Key      string
	Type     Type
	Optional bool
}

type Generic struct {
	typ
	Base       Type
	TypeParams []Type
}

type Ident struct {
	typ
	Parts  []string
	Global bool
}
