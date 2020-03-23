package phpdoc_test

import (
	"reflect"
	"strings"
	"testing"

	"mibk.io/phpdoc"
)

func TestScanner(t *testing.T) {
	const input = `/**
	@param int $foo
	@return string
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
		{phpdoc.Whitespace, " "},
		{phpdoc.Var, "$foo"},
		{phpdoc.Newline, "\n"},
		{phpdoc.Whitespace, "\t"},
		{phpdoc.Tag, "@return"},
		{phpdoc.Whitespace, " "},
		{phpdoc.Ident, "string"},
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
@param DateTime $bar Must be   from this century
@param mixed $foo
 *@return float    Always positive
*/
----
/**
 * @param DateTime $bar Must be   from this century
 * @param mixed $foo
 * @return float Always positive
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
