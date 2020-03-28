package phpdoc

import "strings"

func (doc *PHPDoc) String() string {
	var b strings.Builder
	b.WriteString("/**\n")
	for _, line := range doc.Lines {
		if s := line.String(); s == "" {
			b.WriteString(" *\n")
		} else {
			b.WriteString(" * " + s + "\n")
		}
	}
	b.WriteString(" */\n")
	return b.String()
}

func (txt *TextLine) String() string {
	return txt.Value
}

func (tag *ParamTag) String() string {
	var b strings.Builder
	b.WriteString("@param " + tag.Type.String() + " ")
	if tag.Variadic {
		b.WriteString("...")
	}
	b.WriteRune('$')
	b.WriteString(tag.Var)
	if tag.Desc != "" {
		b.WriteString(" " + tag.Desc)
	}
	return b.String()
}

func (tag *ReturnTag) String() string {
	s := "@return " + tag.Type.String()
	if tag.Desc != "" {
		return s + " " + tag.Desc
	}
	return s
}

func (tag *PropertyTag) String() string {
	var b strings.Builder
	b.WriteString("@property")

	switch {
	case tag.ReadOnly && tag.WriteOnly:
		return "<!invalid property state!>"
	case tag.ReadOnly:
		b.WriteString("-read")
	case tag.WriteOnly:
		b.WriteString("-write")
	}

	b.WriteRune(' ')
	b.WriteString(tag.Type.String())
	if tag.Desc != "" {
		b.WriteString(" " + tag.Desc)
	}
	return b.String()
}

func (tag *OtherTag) String() string {
	s := "@" + tag.Name
	if tag.Desc != "" {
		return s + " " + tag.Desc
	}
	return s
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

func (t *PHPIntersectType) String() string {
	var b strings.Builder
	for i, typ := range t.Types {
		if i > 0 {
			b.WriteRune('&')
		}
		b.WriteString(typ.String())
	}
	return b.String()
}

func (t *PHPParenType) String() string {
	return "(" + t.Type.String() + ")"
}

func (t *PHPArrayType) String() string {
	return t.Elem.String() + "[]"
}

func (t *PHPGenericType) String() string {
	var b strings.Builder
	b.WriteString(t.Base.String())
	b.WriteRune('<')
	for i, typ := range t.Generics {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(typ.String())
	}
	b.WriteRune('>')
	return b.String()
}

func (t *PHPIdentType) String() string {
	if t.Nullable {
		return "?" + t.Name.String()
	}
	return t.Name.String()
}

func (id *PHPIdent) String() string {
	var b strings.Builder
	for i, part := range id.Parts {
		if i > 0 || id.Global {
			b.WriteRune('\\')
		}
		b.WriteString(part)
	}
	return b.String()
}