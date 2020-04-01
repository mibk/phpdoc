package phpdoc_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mibk.io/phpdoc"
)

func TestScanner(t *testing.T) {
	const input = `/**
	@param (\Traversable&\Countable)|array{11 :int} $map
	@param int|null ...$_0_žluťoučký_9
	* @return string[]|array<string, ?string>
*/`

	sc := phpdoc.NewScanner(strings.NewReader(input))

	var got []phpdoc.Token
	for {
		tok := sc.Next()
		got = append(got, tok)
		if tok.Type == phpdoc.EOF {
			break
		}
	}

	want := []phpdoc.Token{
		{phpdoc.OpenDoc, "/**"},
		{phpdoc.Newline, "\n"},
		{phpdoc.Whitespace, "\t"},
		{phpdoc.Tag, "@param"},
		{phpdoc.Whitespace, " "},
		{phpdoc.Lparen, "("},
		{phpdoc.Backslash, "\\"},
		{phpdoc.Ident, "Traversable"},
		{phpdoc.And, "&"},
		{phpdoc.Backslash, "\\"},
		{phpdoc.Ident, "Countable"},
		{phpdoc.Rparen, ")"},
		{phpdoc.Or, "|"},
		{phpdoc.Array, "array"},
		{phpdoc.Lbrace, "{"},
		{phpdoc.Decimal, "11"},
		{phpdoc.Whitespace, " "},
		{phpdoc.Colon, ":"},
		{phpdoc.Ident, "int"},
		{phpdoc.Rbrace, "}"},
		{phpdoc.Whitespace, " "},
		{phpdoc.Var, "$map"},
		{phpdoc.Newline, "\n"},
		{phpdoc.Whitespace, "\t"},
		{phpdoc.Tag, "@param"},
		{phpdoc.Whitespace, " "},
		{phpdoc.Ident, "int"},
		{phpdoc.Or, "|"},
		{phpdoc.Ident, "null"},
		{phpdoc.Whitespace, " "},
		{phpdoc.Ellipsis, "..."},
		{phpdoc.Var, "$_0_žluťoučký_9"},
		{phpdoc.Newline, "\n"},
		{phpdoc.Whitespace, "\t"},
		{phpdoc.Asterisk, "*"},
		{phpdoc.Whitespace, " "},
		{phpdoc.Tag, "@return"},
		{phpdoc.Whitespace, " "},
		{phpdoc.Ident, "string"},
		{phpdoc.Lbrack, "["},
		{phpdoc.Rbrack, "]"},
		{phpdoc.Or, "|"},
		{phpdoc.Array, "array"},
		{phpdoc.Lt, "<"},
		{phpdoc.Ident, "string"},
		{phpdoc.Comma, ","},
		{phpdoc.Whitespace, " "},
		{phpdoc.Query, "?"},
		{phpdoc.Ident, "string"},
		{phpdoc.Gt, ">"},
		{phpdoc.Newline, "\n"},
		{phpdoc.CloseDoc, `*/`},
		{phpdoc.EOF, ""},
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("tokens don't match (-got +want)\n%s", diff)
	}
}

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
	t.Helper()
	sc := phpdoc.NewScanner(strings.NewReader(input))
	p := phpdoc.NewParser(sc)
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
		union      = phpdoc.PHPUnionType
		intersect  = phpdoc.PHPIntersectType
		array      = phpdoc.PHPArrayType
		parens     = phpdoc.PHPParenType
		nullable   = phpdoc.PHPNullableType
		arrayShape = phpdoc.PHPArrayShapeType
		arrayElem  = phpdoc.PHPArrayElem
		generic    = phpdoc.PHPGenericType
		ident      = phpdoc.PHPIdentType
	)

	types := func(types ...phpdoc.PHPType) []phpdoc.PHPType { return types }
	parts := func(parts ...string) []string { return parts }

	tests := []struct {
		typ  string
		want phpdoc.PHPType
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
			want: &generic{Base: new(arrayShape), Generics: types(
				&ident{Parts: parts("string")},
				&array{Elem: &nullable{Type: &generic{
					Base: new(arrayShape),
					Generics: types(
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
				Generics: types(&ident{Parts: parts("T")}),
			},
		},
	}

	for _, tt := range tests {
		sc := phpdoc.NewScanner(strings.NewReader(tt.typ))
		p := phpdoc.NewParser(sc)

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
