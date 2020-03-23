package phpdoc_test

import (
	"reflect"
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

func TestParser(t *testing.T) {
	const input = `/**
@param string $bar
@return float
*/
`

	sc := phpdoc.NewScanner([]byte(input))
	p := phpdoc.NewParser(sc)
	doc, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	if got := doc.String(); got != input {
		t.Errorf("\n got: %q\nwant: %q", got, input)
	}
}
