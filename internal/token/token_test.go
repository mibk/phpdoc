package token_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"mibk.io/phpdoc/internal/token"
)

func TestScanner(t *testing.T) {
	const input = `/**
	@param (\Traversable&\Countable)|array{11 :int} $map
	@param int|null ...$_0_žluťoučký_9
	* @return string[]|array<string, ?string>
*/`

	sc := token.NewScanner(strings.NewReader(input))

	var got []token.Token
	for {
		tok := sc.Next()
		got = append(got, tok)
		if tok.Type == token.EOF {
			break
		}
	}

	want := []token.Token{
		{token.OpenDoc, "/**"},
		{token.Newline, "\n"},
		{token.Whitespace, "\t"},
		{token.TagName, "@param"},
		{token.Whitespace, " "},
		{token.Lparen, "("},
		{token.Backslash, "\\"},
		{token.Ident, "Traversable"},
		{token.And, "&"},
		{token.Backslash, "\\"},
		{token.Ident, "Countable"},
		{token.Rparen, ")"},
		{token.Or, "|"},
		{token.Array, "array"},
		{token.Lbrace, "{"},
		{token.Decimal, "11"},
		{token.Whitespace, " "},
		{token.Colon, ":"},
		{token.Ident, "int"},
		{token.Rbrace, "}"},
		{token.Whitespace, " "},
		{token.VarName, "$map"},
		{token.Newline, "\n"},
		{token.Whitespace, "\t"},
		{token.TagName, "@param"},
		{token.Whitespace, " "},
		{token.Ident, "int"},
		{token.Or, "|"},
		{token.Ident, "null"},
		{token.Whitespace, " "},
		{token.Ellipsis, "..."},
		{token.VarName, "$_0_žluťoučký_9"},
		{token.Newline, "\n"},
		{token.Whitespace, "\t"},
		{token.Asterisk, "*"},
		{token.Whitespace, " "},
		{token.TagName, "@return"},
		{token.Whitespace, " "},
		{token.Ident, "string"},
		{token.Lbrack, "["},
		{token.Rbrack, "]"},
		{token.Or, "|"},
		{token.Array, "array"},
		{token.Lt, "<"},
		{token.Ident, "string"},
		{token.Comma, ","},
		{token.Whitespace, " "},
		{token.Qmark, "?"},
		{token.Ident, "string"},
		{token.Gt, ">"},
		{token.Newline, "\n"},
		{token.CloseDoc, `*/`},
		{token.EOF, ""},
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("tokens don't match (-got +want)\n%s", diff)
	}
}
