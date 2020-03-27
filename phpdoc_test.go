package phpdoc_test

import (
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
 * @param string ...$bar
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
 * @param DateTime|string|null $bar Must be   from this century
 * @param mixed $foo
 * @return float Always positive
 */
`},
	{"tags and text", `
/**
This function does
* * this and
* * that.

 * @author   Jack
It's deprecated now.

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
 * It's deprecated now.
 *
 * @deprecated Don't use
 * @return bool
 */
`},
	{"arrays", `
/**
@param int [ ] $arr
@return string|string[]
*/
----
/**
 * @param int[] $arr
 * @return string|string[]
 */
`},
	{"generics", `
/**
@param array < string, array<string, int > []> $arr
*/
----
/**
 * @param array<string, array<string, int>[]> $arr
 */
`},
	{"intersection", `
/**
@param  Traversable&Countable $map
*/
----
/**
 * @param Traversable&Countable $map
 */
`},
	{"parentheses", `
/**
@param  ( int |float )[] $num
*/
----
/**
 * @param (int|float)[] $num
 */
`},
	{"qualified names", `
/**
@param  \Foo\ Bar \DateTime $a
@param   Other\DateTime     $b
@return \ Traversable
*/
----
/**
 * @param \Foo\Bar\DateTime $a
 * @param Other\DateTime $b
 * @return \Traversable
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
 * @property \Foo $a
 * @property-read array<int, string> $b
 * @property-write int[] $c
 */
`},
}

func TestParser(t *testing.T) {
	for _, tt := range parseTests {
		t.Run(tt.name, func(t *testing.T) {
			s := strings.Split(tt.test, "----\n")
			if len(s) != 2 {
				t.Fatal("invalid test format")
			}

			input, want := strings.TrimSpace(s[0]), s[1]
			parserTestCase(t, input, want)
		})
	}
}

func parserTestCase(t *testing.T, input, want string) {
	t.Helper()
	sc := phpdoc.NewScanner([]byte(input))
	p := phpdoc.NewParser(sc)
	doc, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	if got := doc.String(); got != want {
		t.Errorf("\n got: %q\nwant: %q", got, want)
	}
}
