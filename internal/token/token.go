package token

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

type Pos struct {
	Line, Column int
}

func (p Pos) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

type Token struct {
	Type Type
	Text string
	Pos  Pos
}

func (t Token) String() string {
	if symbolStart < t.Type && t.Type < symbolEnd || t.Type <= Newline {
		return t.Type.String()
	}
	return fmt.Sprintf("%v(%q)", t.Type, t.Text)
}

//go:generate stringer -type Type -linecomment

type Type int

const (
	EOF     Type = iota
	Newline      // \n
	Whitespace

	symbolStart
	OpenDoc  // /**
	CloseDoc // */

	Asterisk  // *
	Backslash // \
	Qmark     // ?
	Lparen    // (
	Rparen    // )
	Lbrack    // [
	Rbrack    // ]
	Lbrace    // {
	Rbrace    // }
	Lt        // <
	Gt        // >
	Comma     // ,
	Colon     // :
	Ellipsis  // ...
	Or        // |
	And       // &
	symbolEnd

	This     // $this
	Array    // array
	Callable // callable

	Ident
	TagName
	VarName

	Decimal

	Other
)

const eof = -1

type Scanner struct {
	r    *bufio.Reader
	done bool

	line, col   int
	lastLineLen int
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		r:    bufio.NewReader(r),
		line: 1,
		col:  1,
	}
}

func (sc *Scanner) Next() Token {
	pos := Pos{Line: sc.line, Column: sc.col}
	tok := sc.scanAny()
	if typ := tok.Type; symbolStart < typ && typ < symbolEnd {
		tok.Text = typ.String()
	}
	tok.Pos = pos
	return tok
}

func (sc *Scanner) read() rune {
	if sc.done {
		return eof
	}
	r, _, err := sc.r.ReadRune()
	if err != nil {
		sc.done = true
		return eof
	}
	if r == '\n' {
		sc.line++
		sc.lastLineLen, sc.col = sc.col, 1
	} else {
		sc.col++
	}
	return r
}

func (sc *Scanner) unread() {
	if sc.done {
		return
	}
	if err := sc.r.UnreadRune(); err != nil {
		// UnreadRune returns an error only on invalid use.
		panic(err)
	}
	sc.col--
	if sc.col == 0 {
		sc.col = sc.lastLineLen
		sc.line--
	}
}

func (sc *Scanner) peek() rune {
	r := sc.read()
	sc.unread()
	return r
}

func (sc *Scanner) scanAny() Token {
	switch r := sc.read(); r {
	case eof:
		return Token{Type: EOF}
	case '/':
		if sc.peek() == '*' {
			sc.read()
			return sc.scanOpenDoc()
		}
		return sc.scanOther("/")
	case '@':
		return sc.scanTag()
	case '$':
		return sc.scanVar()
	case '*':
		if sc.peek() == '/' {
			sc.read()
			return Token{Type: CloseDoc, Text: "*/"}
		}
		return Token{Type: Asterisk}
	case '\\':
		return Token{Type: Backslash}
	case '?':
		return Token{Type: Qmark}
	case '(':
		return Token{Type: Lparen}
	case ')':
		return Token{Type: Rparen}
	case '[':
		return Token{Type: Lbrack}
	case ']':
		return Token{Type: Rbrack}
	case '{':
		return Token{Type: Lbrace}
	case '}':
		return Token{Type: Rbrace}
	case '<':
		return Token{Type: Lt}
	case '>':
		return Token{Type: Gt}
	case ',':
		return Token{Type: Comma}
	case ':':
		return Token{Type: Colon}
	case '.':
		if sc.peek() == '.' {
			if sc.read(); sc.peek() == '.' {
				sc.read()
				return Token{Type: Ellipsis}
			}
			return sc.scanOther("..")
		}
		return sc.scanOther(".")
	case '|':
		return Token{Type: Or}
	case '&':
		return Token{Type: And}
	case '\n':
		return Token{Type: Newline, Text: string(r)}
	case ' ', '\t':
		return sc.scanWhitespace(r)
	default:
		if isDigit(r) {
			return sc.scanDecimal(r)
		}
		sc.unread()
		return sc.scanOther("")
	}
}

func (sc *Scanner) scanOpenDoc() Token {
	r := sc.read()
	if r == '*' {
		return Token{Type: OpenDoc, Text: "/**"}
	}
	sc.unread()
	return sc.scanOther("/*")
}

func (sc *Scanner) scanTag() Token {
	id := sc.scanTagName()
	if id == "" {
		return sc.scanOther("@")
	}
	return Token{Type: TagName, Text: "@" + id}
}

func (sc *Scanner) scanTagName() string {
	var b strings.Builder
	for {
		switch r := sc.read(); {
		case r == '-' || r >= 'a' && r <= 'z':
			b.WriteRune(r)
		default:
			sc.unread()
			return b.String()
		}
	}
}

func (sc *Scanner) scanVar() Token {
	switch id := sc.scanIdentName(); {
	case id == "", strings.ContainsRune(id, '-'):
		return sc.scanOther("$" + id)
	case id == "this":
		return Token{Type: This, Text: "$this"}
	default:
		return Token{Type: VarName, Text: "$" + id}
	}
}

func (sc *Scanner) scanIdentName() string {
	var b strings.Builder
	for {
		switch r := sc.read(); {
		case r == '_' || r == '-' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= utf8.RuneSelf:
			// A dash (-) actually isn't allowed in a PHP identifier,
			// but it's used in meta-types (e.g. class-name). See
			// https://psalm.dev/docs/annotating_code/type_syntax/scalar_types/
			// for more information.
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			if b.Len() > 0 {
				b.WriteRune(r)
				continue
			}
			fallthrough
		default:
			sc.unread()
			return b.String()
		}
	}
}

func (sc *Scanner) scanDecimal(r rune) Token {
	var b strings.Builder
	b.WriteRune(r)
	for isDigit(sc.peek()) {
		b.WriteRune(sc.read())
	}
	return Token{Type: Decimal, Text: b.String()}
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func (sc *Scanner) scanWhitespace(init rune) Token {
	var b strings.Builder
	b.WriteRune(init)
	for {
		switch r := sc.read(); r {
		case ' ', '\t':
			b.WriteRune(r)
		default:
			sc.unread()
			return Token{Type: Whitespace, Text: b.String()}
		}
	}
}

func (sc *Scanner) scanOther(init string) Token {
	if init == "" {
		switch id := sc.scanIdentName(); id {
		case "":
			break
		case "array":
			return Token{Type: Array, Text: id}
		case "callable":
			return Token{Type: Callable, Text: id}
		default:
			return Token{Type: Ident, Text: id}
		}
	}

	var b strings.Builder
	b.WriteString(init)
	for {
		switch r := sc.read(); r {
		default:
			b.WriteRune(r)
		case '\n', '@', '$', '*', eof:
			sc.unread()
			return Token{Type: Other, Text: b.String()}
		}
	}
}
