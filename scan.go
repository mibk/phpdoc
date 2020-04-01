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
	EOF TokenType = iota
	Error
	Newline // \n
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

	Ident // baz
	Tag   // @foo
	Var   // $bar

	Decimal // 1

	Other
)

const eof = -1

type Scanner struct {
	r *bufio.Reader

	err    error
	buf    bytes.Buffer
	last   rune
	peeked rune
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

func (sc *Scanner) Next() Token {
	if sc.err != nil {
		// TODO type Error
		return Token{Type: EOF}
	}
	return sc.lexAny()
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
		sc.err = err
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

func (sc *Scanner) lexAny() Token {
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
	id := sc.lexTag()
	if id == "" {
		return sc.scanOther("@")
	}
	return Token{Type: Tag, Text: "@" + id}
}

func (sc *Scanner) lexTag() string {
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
	id := sc.lexIdent()
	if id == "" {
		return sc.scanOther("$")
	}
	return Token{Type: Var, Text: "$" + id}
}

func (sc *Scanner) lexIdent() string {
	var b strings.Builder
	for {
		switch r := sc.next(); {
		case r == '_' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= utf8.RuneSelf:
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
		switch id := sc.lexIdent(); id {
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
