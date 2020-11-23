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
	switch {
	case t.Type == EOF, t.Type == Newline,
		symbolStart < t.Type && t.Type < symbolEnd,
		keywordStart < t.Type && t.Type < keywordEnd:
		return t.Type.String()
	default:
		return fmt.Sprintf("%v(%q)", t.Type, t.Text)
	}
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

	keywordStart
	This     // $this
	Array    // array
	Callable // callable
	Static   // static
	keywordEnd

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

func (s *Scanner) Next() Token {
	pos := Pos{Line: s.line, Column: s.col}
	tok := s.scanAny()
	if typ := tok.Type; symbolStart < typ && typ < symbolEnd {
		tok.Text = typ.String()
	}
	tok.Pos = pos
	return tok
}

func (s *Scanner) read() rune {
	if s.done {
		return eof
	}
	r, _, err := s.r.ReadRune()
	if err != nil {
		s.done = true
		return eof
	}
	if r == '\n' {
		s.line++
		s.lastLineLen, s.col = s.col, 1
	} else {
		s.col++
	}
	return r
}

func (s *Scanner) unread() {
	if s.done {
		return
	}
	if err := s.r.UnreadRune(); err != nil {
		// UnreadRune returns an error only on invalid use.
		panic(err)
	}
	s.col--
	if s.col == 0 {
		s.col = s.lastLineLen
		s.line--
	}
}

func (s *Scanner) peek() rune {
	r := s.read()
	s.unread()
	return r
}

func (s *Scanner) scanAny() Token {
	switch r := s.read(); r {
	case eof:
		return Token{Type: EOF}
	case '/':
		if s.peek() == '*' {
			s.read()
			return s.scanOpenDoc()
		}
		return s.scanOther("/")
	case '@':
		return s.scanTag()
	case '$':
		return s.scanVar()
	case '*':
		if s.peek() == '/' {
			s.read()
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
		if s.peek() == '.' {
			if s.read(); s.peek() == '.' {
				s.read()
				return Token{Type: Ellipsis}
			}
			return s.scanOther("..")
		}
		return s.scanOther(".")
	case '|':
		return Token{Type: Or}
	case '&':
		return Token{Type: And}
	case '\n':
		return Token{Type: Newline, Text: string(r)}
	case ' ', '\t':
		return s.scanWhitespace(r)
	default:
		if isDigit(r) {
			return s.scanDecimal(r)
		}
		s.unread()
		return s.scanOther("")
	}
}

func (s *Scanner) scanOpenDoc() Token {
	r := s.read()
	if r == '*' {
		return Token{Type: OpenDoc, Text: "/**"}
	}
	s.unread()
	return s.scanOther("/*")
}

func (s *Scanner) scanTag() Token {
	id := s.scanTagName()
	if id == "" {
		return s.scanOther("@")
	}
	return Token{Type: TagName, Text: "@" + id}
}

func (s *Scanner) scanTagName() string {
	var b strings.Builder
	for {
		switch r := s.read(); {
		case r == '-' || r >= 'a' && r <= 'z':
			b.WriteRune(r)
		default:
			s.unread()
			return b.String()
		}
	}
}

func (s *Scanner) scanVar() Token {
	switch id := s.scanIdentName(); {
	case id == "", strings.ContainsRune(id, '-'):
		return s.scanOther("$" + id)
	case id == "this":
		return Token{Type: This, Text: "$this"}
	default:
		return Token{Type: VarName, Text: "$" + id}
	}
}

func (s *Scanner) scanIdentName() string {
	var b strings.Builder
	for {
		switch r := s.read(); {
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
			s.unread()
			return b.String()
		}
	}
}

func (s *Scanner) scanDecimal(r rune) Token {
	var b strings.Builder
	b.WriteRune(r)
	for isDigit(s.peek()) {
		b.WriteRune(s.read())
	}
	return Token{Type: Decimal, Text: b.String()}
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func (s *Scanner) scanWhitespace(init rune) Token {
	var b strings.Builder
	b.WriteRune(init)
	for {
		switch r := s.read(); r {
		case ' ', '\t':
			b.WriteRune(r)
		default:
			s.unread()
			return Token{Type: Whitespace, Text: b.String()}
		}
	}
}

func (s *Scanner) scanOther(init string) Token {
	if init == "" {
		switch id := s.scanIdentName(); id {
		case "":
			break
		case "array":
			return Token{Type: Array, Text: id}
		case "callable":
			return Token{Type: Callable, Text: id}
		case "static":
			return Token{Type: Static, Text: id}
		default:
			return Token{Type: Ident, Text: id}
		}
	}

	var b strings.Builder
	b.WriteString(init)
	for {
		switch r := s.read(); r {
		default:
			b.WriteRune(r)
		case '\n', '@', '$', '*', eof:
			s.unread()
			return Token{Type: Other, Text: b.String()}
		}
	}
}
