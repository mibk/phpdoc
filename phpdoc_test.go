package phpdoc_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mibk.io/phpdoc"
	"mibk.io/phpdoc/phptype"
)

var parseTests = []struct {
	name string
	test string
}{
	{"basic", `
/**
	@param string ... $bar
	@return ? float
*/
----
/**
 * @param  string ...$bar
 * @return ?float
 */
`},
	{"oneline", `
  /**@var \ DateTime $date    */
----
  /** @var \DateTime $date */
`},
	{"single line", `
  /**
@var \ Traversable*/
----
  /**
   * @var \Traversable
   */
`},
	{"more params", `
   /**
	@author   Name <not known>
@param DateTime | string|null $bar Must be   from this century
@param mixed $foo
 *@return float    Always positive
*/
----
   /**
    * @author Name <not known>
    * @param  DateTime|string|null $bar Must be   from this century
    * @param  mixed                $foo
    * @return float                Always positive
    */
`},
	{"tags and text", `
/**
This function does
* * this and
* * that.

 * @author   Jack
It's	deprecated now.

@deprecated Don't use
@return bool
*/
----
/**
 * This function does
 * * this and
 * * that.
 *
 * @author Jack
 * It's	deprecated now.
 *
 * @deprecated Don't use
 * @return     bool
 */
`},
	{"properties", `
	/**
@property  \ Foo $a
@property-read    array<int,string>    $b
@property-write int [] $c
@property array    {0 :int  ,foo?:\ Foo }$d
*/
----
	/**
	 * @property       \Foo                      $a
	 * @property-read  array<int, string>        $b
	 * @property-write int[]                     $c
	 * @property       array{0: int, foo?: \Foo} $d
	 */
`},
	{"template", `
/**
@template    T foo
@template  U of \ Traversable bar
@template   WW as \ Countable */
----
/**
 * @template T                 foo
 * @template U of \Traversable bar
 * @template WW of \Countable
 */
`},
}

func TestPrinting(t *testing.T) {
	for _, tt := range parseTests {
		t.Run(tt.name, func(t *testing.T) {
			s := strings.Split(tt.test, "----\n")
			if len(s) != 2 {
				t.Fatal("invalid test format")
			}

			input, want := s[0], s[1]
			printerTestCase(t, input, want)
		})
	}
}

func printerTestCase(t *testing.T, input, want string) {
	p := phpdoc.NewParser(strings.NewReader(input))
	doc, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	got := new(strings.Builder)
	if err := phpdoc.Fprint(got, doc); err != nil {
		t.Fatalf("printing: unexpected err: %v", err)
	}
	if got.String() != want {
		t.Errorf("\n got: %q\nwant: %q", got, want)
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
		p := phpdoc.NewParser(strings.NewReader(tt.typ))

		got, err := p.ParseType()
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
			`line:2:6: expecting Ident, found EOF`,
		},
		{
			"/**\n@param array<int  string>",
			`line:2:19: expecting >, found Ident("string")`,
		},
		{
			"/**\n@param array<int, >",
			`line:2:19: expecting Ident, found >`,
		},
	}

	for _, tt := range tests {
		p := phpdoc.NewParser(strings.NewReader(tt.doc))
		doc, err := p.Parse()
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
