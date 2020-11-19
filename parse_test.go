package phpdoc_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mibk.io/phpdoc"
	"mibk.io/phpdoc/phptype"
)

func TestParsingDoc(t *testing.T) {
	lines := func(lines ...phpdoc.Line) []phpdoc.Line { return lines }
	typ := func(name string) phptype.Type { return &phptype.Ident{Parts: []string{name}} }

	tests := []struct {
		doc  string
		want interface{}
	}{
		{
			doc:  `/** */`,
			want: &phpdoc.Block{PreferOneline: true},
		},
		{
			doc: `/** Foo  $xx. */`,
			want: &phpdoc.Block{
				Lines:         lines(&phpdoc.TextLine{Value: "Foo  $xx."}),
				PreferOneline: true,
			},
		},
		{
			doc: `/** @var Foo $bar Baz   x*/`,
			want: &phpdoc.Block{
				Lines:         lines(&phpdoc.VarTag{Type: typ("Foo"), Var: "bar", Desc: "Baz   x"}),
				PreferOneline: true,
			},
		},
		{
			doc: `/** @property int $id {primary} */`,
			want: &phpdoc.Block{
				Lines:         lines(&phpdoc.PropertyTag{Type: typ("int"), Var: "id", Desc: "{primary}"}),
				PreferOneline: true,
			},
		},
	}

	for _, tt := range tests {
		got, err := phpdoc.Parse(strings.NewReader(tt.doc))
		if err != nil {
			t.Fatalf("%q: unexpected err: %v", tt.doc, err)
		}

		allowUnexportedFields := cmp.Exporter(func(reflect.Type) bool { return true })
		if diff := cmp.Diff(got, tt.want, allowUnexportedFields); diff != "" {
			t.Errorf("%q: docs don't match (-got +want)\n%s", tt.doc, diff)
		}
	}
}

func TestParsingTypes(t *testing.T) {
	type (
		union      = phptype.Union
		intersect  = phptype.Intersect
		array      = phptype.Array
		parens     = phptype.Paren
		nullable   = phptype.Nullable
		arrayShape = phptype.ArrayShape
		arrayElem  = phptype.ArrayElem
		generic    = phptype.Generic
		ident      = phptype.Ident
	)

	types := func(types ...phptype.Type) []phptype.Type { return types }
	parts := func(parts ...string) []string { return parts }

	tests := []struct {
		typ  string
		want phptype.Type
	}{
		{
			typ:  `? float`,
			want: &nullable{Type: &ident{Parts: parts("float")}},
		},
		{
			typ:  `int [ ] []`,
			want: &array{Elem: &array{Elem: &ident{Parts: parts("int")}}},
		},
		{
			typ: `array < string, ?array<string, int > []>`,
			want: &generic{Base: new(arrayShape), TypeParams: types(
				&ident{Parts: parts("string")},
				&array{Elem: &nullable{Type: &generic{
					Base: new(arrayShape),
					TypeParams: types(
						&ident{Parts: parts("string")},
						&ident{Parts: parts("int")},
					),
				}}},
			)},
		},
		{
			typ: `Traversable &Countable`,
			want: &intersect{Types: types(
				&ident{Parts: parts("Traversable")},
				&ident{Parts: parts("Countable")},
			)},
		},
		{
			typ: `( int |float )[]`,
			want: &array{Elem: &parens{Type: &union{Types: types(
				&ident{Parts: parts("int")},
				&ident{Parts: parts("float")},
			)}}},
		},
		{
			typ:  `\Foo\ Bar \DateTime`,
			want: &ident{Parts: parts("Foo", "Bar", "DateTime"), Global: true},
		},
		{
			typ:  `Other\DateTime`,
			want: &ident{Parts: parts("Other", "DateTime")},
		},
		{
			typ:  `\ Traversable`,
			want: &ident{Parts: parts("Traversable"), Global: true},
		},
		{
			typ: `array{20?:int, foo :string | \ DateTime}`,
			want: &arrayShape{Elems: []*arrayElem{
				{Key: "20", Type: &ident{Parts: parts("int")}, Optional: true},
				{Key: "foo", Type: &union{Types: types(
					&ident{Parts: parts("string")},
					&ident{Parts: parts("DateTime"), Global: true},
				)}},
			}},
		},
		{
			typ: `class-string<T>`,
			want: &generic{Base: &ident{Parts: parts("class-string")},
				TypeParams: types(&ident{Parts: parts("T")}),
			},
		},
	}

	for _, tt := range tests {
		got, err := phpdoc.ParseType(strings.NewReader(tt.typ))
		if err != nil {
			t.Fatalf("%q: unexpected err: %v", tt.typ, err)
		}

		allowUnexportedFields := cmp.Exporter(func(reflect.Type) bool { return true })
		if diff := cmp.Diff(got, tt.want, allowUnexportedFields); diff != "" {
			t.Errorf("%q: types don't match (-got +want)\n%s", tt.typ, diff)
		}
	}
}

func TestSyntaxErrors(t *testing.T) {
	tests := []struct {
		doc     string
		wantErr string
	}{
		{
			"/**",
			`line:1:4: expecting */, found EOF`,
		},
		{
			"/**\n@param",
			`line:2:6: expecting ( or basic type, found EOF`,
		},
		{
			"/**\n@param array<int  string>",
			`line:2:19: expecting >, found Ident("string")`,
		},
		{
			"/**\n@param array<int, >",
			`line:2:19: expecting ( or basic type, found >`,
		},
		{
			"/**\n@param callable(int&)",
			`line:2:21: expecting VarName, found )`,
		},
		{
			"/**@param int*/",
			`line:1:14: expecting VarName, found */`,
		},
		{
			"/**@param string $this*/",
			`line:1:18: expecting VarName, found $this("$this")`,
		},
	}

	for _, tt := range tests {
		doc, err := phpdoc.Parse(strings.NewReader(tt.doc))
		errStr := "<nil>"
		if err != nil {
			if doc != nil {
				t.Fatalf("%q: got %+v on err", tt.doc, doc)
			}
			errStr = err.Error()
		}
		if errStr != tt.wantErr {
			t.Errorf("%q:\n got %s\nwant %s", tt.doc, errStr, tt.wantErr)
		}
	}
}
