package phpdoc_test

import (
	"reflect"
	"strings"
	"testing"

	"mibk.io/phpdoc"
)

func TestScanner(t *testing.T) {
	const input = `/**
	@param int|null $foo
	* @return string[]
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
		{phpdoc.Ident, "int"},
		{phpdoc.Union, "|"},
		{phpdoc.Ident, "null"},
		{phpdoc.Whitespace, " "},
		{phpdoc.Var, "$foo"},
		{phpdoc.Newline, "\n"},
		{phpdoc.Whitespace, "\t"},
		{phpdoc.Asterisk, "*"},
		{phpdoc.Whitespace, " "},
		{phpdoc.Tag, "@return"},
		{phpdoc.Whitespace, " "},
		{phpdoc.Ident, "string"},
		{phpdoc.OpenBrack, "["},
		{phpdoc.CloseBrack, "]"},
		{phpdoc.Newline, "\n"},
		{phpdoc.CloseDoc, `*/`},
		{phpdoc.EOF, ""},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("\n got: %v\nwant: %v", got, want)
	}
}

var parseTests = []struct {
	name string
	test string
}{
	{"basic", `
/**
	@param string $bar
	@return float
*/
----
/**
 * @param string $bar
 * @return float
 */
`},
	{"more params", `
/**
	@author   Name <not known>
@param DateTime|string|null $bar Must be   from this century
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
@param int[] $arr
@return string|string[]
*/
----
/**
 * @param int[] $arr
 * @return string|string[]
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
