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
	@param (\Traversable&\Countable)|array $map
	@param int|null ...$_0_žluťoučký_9
	* @return string[]|array<string, ?string>
*/`

	sc := phpdoc.NewScanner([]byte(input))

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
		{phpdoc.OpenParen, "("},
		{phpdoc.Backslash, "\\"},
		{phpdoc.Ident, "Traversable"},
		{phpdoc.Intersect, "&"},
		{phpdoc.Backslash, "\\"},
		{phpdoc.Ident, "Countable"},
		{phpdoc.CloseParen, ")"},
		{phpdoc.Union, "|"},
		{phpdoc.Ident, "array"},
		{phpdoc.Whitespace, " "},
		{phpdoc.Var, "$map"},
		{phpdoc.Newline, "\n"},
		{phpdoc.Whitespace, "\t"},
		{phpdoc.Tag, "@param"},
		{phpdoc.Whitespace, " "},
		{phpdoc.Ident, "int"},
		{phpdoc.Union, "|"},
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
		{phpdoc.OpenBrack, "["},
		{phpdoc.CloseBrack, "]"},
		{phpdoc.Union, "|"},
		{phpdoc.Ident, "array"},
		{phpdoc.OpenAngle, "<"},
		{phpdoc.Ident, "string"},
		{phpdoc.Comma, ","},
		{phpdoc.Whitespace, " "},
		{phpdoc.Nullable, "?"},
		{phpdoc.Ident, "string"},
		{phpdoc.CloseAngle, ">"},
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
*/
----
/**
 * @property       \Foo               $a
 * @property-read  array<int, string> $b
 * @property-write int[]              $c
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

			input, want := strings.TrimSpace(s[0]), s[1]
			printerTestCase(t, input, want)
		})
	}
}

func printerTestCase(t *testing.T, input, want string) {
	t.Helper()
	sc := phpdoc.NewScanner([]byte(input))
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
		union     = phpdoc.PHPUnionType
		intersect = phpdoc.PHPIntersectType
		array     = phpdoc.PHPArrayType
		parens    = phpdoc.PHPParenType
		generic   = phpdoc.PHPGenericType
		ident     = phpdoc.PHPIdentType
		name      = phpdoc.PHPIdent
	)

	types := func(types ...phpdoc.PHPType) []phpdoc.PHPType { return types }
	parts := func(parts ...string) []string { return parts }

	tests := []struct {
		typ  string
		want phpdoc.PHPType
	}{
		{
			typ:  `? float`,
			want: &ident{Name: &name{Parts: parts("float")}, Nullable: true},
		},
		{
			typ:  `int [ ] []`,
			want: &array{Elem: &array{Elem: &ident{Name: &name{Parts: parts("int")}}}},
		},
		{
			typ: `array < string, array<string, int > []>`,
			want: &generic{Base: &ident{Name: &name{Parts: parts("array")}}, Generics: types(
				&ident{Name: &name{Parts: parts("string")}},
				&array{Elem: &generic{
					Base: &ident{Name: &name{Parts: parts("array")}},
					Generics: types(
						&ident{Name: &name{Parts: parts("string")}},
						&ident{Name: &name{Parts: parts("int")}},
					),
				}},
			)},
		},
		{
			typ: `Traversable &Countable`,
			want: &intersect{Types: types(
				&ident{Name: &name{Parts: parts("Traversable")}},
				&ident{Name: &name{Parts: parts("Countable")}},
			)},
		},
		{
			typ: `( int |float )[]`,
			want: &array{Elem: &parens{Type: &union{Types: types(
				&ident{Name: &name{Parts: parts("int")}},
				&ident{Name: &name{Parts: parts("float")}},
			)}}},
		},
		{
			typ:  `\Foo\ Bar \DateTime`,
			want: &ident{Name: &name{Parts: parts("Foo", "Bar", "DateTime"), Global: true}},
		},
		{
			typ:  `Other\DateTime`,
			want: &ident{Name: &name{Parts: parts("Other", "DateTime")}},
		},
		{
			typ:  `\ Traversable`,
			want: &ident{Name: &name{Parts: parts("Traversable"), Global: true}},
		},
	}

	for _, tt := range tests {
		sc := phpdoc.NewScanner([]byte(tt.typ))
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
