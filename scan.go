package phpdoc

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
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
	Other
	OpenDoc    // /**
	CloseDoc   // */
	Tag        // @foo
	Var        // $bar
	Nullable   // ?
	OpenParen  // (
	CloseParen // )
	OpenBrack  // [
	CloseBrack // ]
	OpenAngle  // <
	CloseAngle // >
	Comma      // ,
	Ellipsis   // ...
	Union      // |
	Intersect  // &
	Ident      // baz
)

const eof = -1

type Scanner struct {
	r *bufio.Reader

	err    error
	buf    bytes.Buffer
	last   rune
	peeked rune
}

func NewScanner(input []byte) *Scanner {
	return &Scanner{r: bufio.NewReader(bytes.NewReader(input))}
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
	case '?':
		return Token{Type: Nullable, Text: "?"}
	case '(':
		return Token{Type: OpenParen, Text: "("}
	case ')':
		return Token{Type: CloseParen, Text: ")"}
	case '[':
		return Token{Type: OpenBrack, Text: "["}
	case ']':
		return Token{Type: CloseBrack, Text: "]"}
	case '<':
		return Token{Type: OpenAngle, Text: "<"}
	case '>':
		return Token{Type: CloseAngle, Text: ">"}
	case ',':
		return Token{Type: Comma, Text: ","}
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
		return Token{Type: Union, Text: "|"}
	case '&':
		return Token{Type: Intersect, Text: "&"}
	case '\n':
		return Token{Type: Newline, Text: string(r)}
	case ' ', '\t':
		return sc.scanWhitespace(r)
	default:
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
	id := sc.lexIdent()
	if id == "" {
		return sc.scanOther("@")
	}
	return Token{Type: Tag, Text: "@" + id}
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
		case r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		default:
			sc.backup()
			return b.String()
		}
	}
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
		id := sc.lexIdent()
		if id != "" {
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
