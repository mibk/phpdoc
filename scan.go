package phpdoc

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

type Token struct {
	Type TokenType
	Text string
}

func (t Token) String() string {
	return fmt.Sprintf("%v(%q)", t.Type, t.Text)
}

//go:generate stringer -type TokenType

type TokenType int

const (
	EOF     TokenType = iota
	Newline           // \n
	Whitespace
	Asterisk // *

	OpenDoc  // /**
	CloseDoc // */

	Backslash // \
	Query     // ?
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

	Array // array

	Ident   // baz
	TagName // @foo
	VarName // $bar

	Decimal // 1

	Other
)

const eof = -1

type Scanner struct {
	r *bufio.Reader

	buf    bytes.Buffer
	last   rune
	peeked rune
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

func (sc *Scanner) Next() Token {
	return sc.scanAny()
}

func (sc *Scanner) next() rune {
	if sc.peeked != 0 {
		r := sc.peeked
		sc.peeked = 0

		return r
	}

	r, _, err := sc.r.ReadRune()
	if err != nil {
		r = eof
	}
	sc.last = r
	return r
}

func (sc *Scanner) backup() {
	sc.peeked = sc.last
}

func (sc *Scanner) peek() rune {
	r := sc.next()
	sc.backup()
	return r
}

func (sc *Scanner) scanAny() Token {
	switch r := sc.next(); r {
	case eof:
		return Token{Type: EOF}
	case '/':
		if sc.peek() == '*' {
			sc.next()
			return sc.scanOpenDoc()
		}
		return sc.scanOther("/")
	case '@':
		return sc.scanTag()
	case '$':
		return sc.scanVar()
	case '*':
		if sc.peek() == '/' {
			sc.next()
			return Token{Type: CloseDoc, Text: "*/"}
		}
		return Token{Type: Asterisk, Text: "*"}
	case '\\':
		return Token{Type: Backslash, Text: "\\"}
	case '?':
		return Token{Type: Query, Text: "?"}
	case '(':
		return Token{Type: Lparen, Text: "("}
	case ')':
		return Token{Type: Rparen, Text: ")"}
	case '[':
		return Token{Type: Lbrack, Text: "["}
	case ']':
		return Token{Type: Rbrack, Text: "]"}
	case '{':
		return Token{Type: Lbrace, Text: "{"}
	case '}':
		return Token{Type: Rbrace, Text: "}"}
	case '<':
		return Token{Type: Lt, Text: "<"}
	case '>':
		return Token{Type: Gt, Text: ">"}
	case ',':
		return Token{Type: Comma, Text: ","}
	case ':':
		return Token{Type: Colon, Text: ":"}
	case '.':
		if sc.peek() == '.' {
			if sc.next(); sc.peek() == '.' {
				sc.next()
				return Token{Type: Ellipsis, Text: "..."}
			}
			return sc.scanOther("..")
		}
		return sc.scanOther(".")
	case '|':
		return Token{Type: Or, Text: "|"}
	case '&':
		return Token{Type: And, Text: "&"}
	case '\n':
		return Token{Type: Newline, Text: string(r)}
	case ' ', '\t':
		return sc.scanWhitespace(r)
	default:
		if isDigit(r) {
			return sc.scanDecimal(r)
		}
		sc.backup()
		return sc.scanOther("")
	}
}

func (sc *Scanner) scanOpenDoc() Token {
	r := sc.next()
	if r == '*' {
		return Token{Type: OpenDoc, Text: "/**"}
	}
	sc.backup()
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
		switch r := sc.next(); {
		case r == '-' || r >= 'a' && r <= 'z':
			b.WriteRune(r)
		default:
			sc.backup()
			return b.String()
		}
	}
}

func (sc *Scanner) scanVar() Token {
	switch id := sc.scanIdentName(); {
	case id == "", strings.ContainsRune(id, '-'):
		return sc.scanOther("$" + id)
	default:
		return Token{Type: VarName, Text: "$" + id}
	}
}

func (sc *Scanner) scanIdentName() string {
	var b strings.Builder
	for {
		switch r := sc.next(); {
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
			sc.backup()
			return b.String()
		}
	}
}

func (sc *Scanner) scanDecimal(r rune) Token {
	var b strings.Builder
	b.WriteRune(r)
	for isDigit(sc.peek()) {
		b.WriteRune(sc.next())
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
		switch r := sc.next(); r {
		default:
			sc.backup()
			fallthrough
		case eof:
			return Token{Type: Whitespace, Text: b.String()}
		case ' ', '\t':
			b.WriteRune(r)
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
		default:
			return Token{Type: Ident, Text: id}
		}
	}

	var b strings.Builder
	b.WriteString(init)
	for {
		switch r := sc.next(); r {
		case '\n', '@', '$', '*':
			sc.backup()
			fallthrough
		case eof:
			return Token{Type: Other, Text: b.String()}
		default:
			b.WriteRune(r)
		}
	}
}
