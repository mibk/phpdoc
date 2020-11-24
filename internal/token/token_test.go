package token_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mibk.io/phpdoc/internal/token"
)

func pos(posStr string) token.Pos {
	var pos token.Pos
	fmt.Sscanf(posStr, "%d:%d", &pos.Line, &pos.Column)
	return pos
}

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
		{token.OpenDoc, "/**", pos("1:1")},
		{token.Newline, "\n", pos("1:4")},
		{token.Whitespace, "\t", pos("2:1")},
		{token.Tag, "@param", pos("2:2")},
		{token.Whitespace, " ", pos("2:8")},
		{token.Lparen, "(", pos("2:9")},
		{token.Backslash, "\\", pos("2:10")},
		{token.Ident, "Traversable", pos("2:11")},
		{token.And, "&", pos("2:22")},
		{token.Backslash, "\\", pos("2:23")},
		{token.Ident, "Countable", pos("2:24")},
		{token.Rparen, ")", pos("2:33")},
		{token.Or, "|", pos("2:34")},
		{token.Array, "array", pos("2:35")},
		{token.Lbrace, "{", pos("2:40")},
		{token.Int, "11", pos("2:41")},
		{token.Whitespace, " ", pos("2:43")},
		{token.Colon, ":", pos("2:44")},
		{token.Ident, "int", pos("2:45")},
		{token.Rbrace, "}", pos("2:48")},
		{token.Whitespace, " ", pos("2:49")},
		{token.Var, "$map", pos("2:50")},
		{token.Newline, "\n", pos("2:54")},
		{token.Whitespace, "\t", pos("3:1")},
		{token.Tag, "@param", pos("3:2")},
		{token.Whitespace, " ", pos("3:8")},
		{token.Ident, "int", pos("3:9")},
		{token.Or, "|", pos("3:12")},
		{token.Ident, "null", pos("3:13")},
		{token.Whitespace, " ", pos("3:17")},
		{token.Ellipsis, "...", pos("3:18")},
		{token.Var, "$_0_žluťoučký_9", pos("3:21")},
		{token.Newline, "\n", pos("3:36")},
		{token.Whitespace, "\t", pos("4:1")},
		{token.Asterisk, "*", pos("4:2")},
		{token.Whitespace, " ", pos("4:3")},
		{token.Tag, "@return", pos("4:4")},
		{token.Whitespace, " ", pos("4:11")},
		{token.Ident, "string", pos("4:12")},
		{token.Lbrack, "[", pos("4:18")},
		{token.Rbrack, "]", pos("4:19")},
		{token.Or, "|", pos("4:20")},
		{token.Array, "array", pos("4:21")},
		{token.Lt, "<", pos("4:26")},
		{token.Ident, "string", pos("4:27")},
		{token.Comma, ",", pos("4:33")},
		{token.Whitespace, " ", pos("4:34")},
		{token.Qmark, "?", pos("4:35")},
		{token.Ident, "string", pos("4:36")},
		{token.Gt, ">", pos("4:42")},
		{token.Newline, "\n", pos("4:43")},
		{token.CloseDoc, `*/`, pos("5:1")},
		{token.EOF, "", pos("5:3")},
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("tokens don't match: (-got +want)\n%s", diff)
	}
}
