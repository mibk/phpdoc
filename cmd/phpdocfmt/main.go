// PHPDocfmt formats PHPDoc comments in PHP files.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"mibk.io/php/token"
	"mibk.io/phpdoc"
)

var inPlace = flag.Bool("w", false, "write to file")

func main() {
	flag.Parse()
	log.SetPrefix("phpdocfmt: ")
	log.SetFlags(0)

	if flag.NArg() == 0 {
		if *inPlace {
			log.Fatal("cannot use -w with standard input")
		}
		if err := formatDocs("<stdin>", os.Stdout, os.Stdin); err != nil {
			log.Fatal(err)
		}
		return
	}

	for _, filename := range flag.Args() {
		f, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		fi, err := f.Stat()
		if err != nil {
			log.Fatal(err)
		}
		perm := fi.Mode().Perm()

		buf := new(bytes.Buffer)
		if err := formatDocs(filename, buf, f); err != nil {
			log.Println(err)
		}
		f.Close()

		if *inPlace {
			// TODO: Make backup file?
			if err := ioutil.WriteFile(filename, buf.Bytes(), perm); err != nil {
				log.Fatal(err)
			}
		} else {
			if _, err := io.Copy(os.Stdout, buf); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func formatDocs(filename string, out io.Writer, in io.Reader) error {
	scan := token.NewScanner(in)

	w := &stickyErrWriter{w: out}
	var ws string
	var doc *phpdoc.Block
Loop:
	for {
		tok := scan.Next()
		if doc != nil {
			ws = ""
			if tok.Type == token.Whitespace {
				i := strings.LastIndexByte(tok.Text, '\n')
				doc.Indent = tok.Text[i+1:]
			}
			if err := phpdoc.Fprint(w, doc); err != nil {
				return fmt.Errorf("%s: printing doc: %v", filename, err)
			}
			io.WriteString(w, doc.Indent)
			doc = nil
			if tok.Type == token.Whitespace {
				continue
			}
		}
		if tok.Type == token.DocComment {
			var err error
			doc, err = phpdoc.Parse(strings.NewReader(tok.Text))
			if err == nil {
				continue
			}
			pos := Pos{Line: tok.Pos.Line, Column: tok.Pos.Column}
			if se, ok := err.(*phpdoc.SyntaxError); ok {
				pos = pos.Add(Pos{Line: se.Line, Column: se.Column})
				err = se.Err
			}
			return fmt.Errorf("%s:%v: %v", filename, pos, err)
		}
		if ws != "" {
			io.WriteString(w, ws)
			ws = ""
		}
		switch tok.Type {
		case token.EOF:
			break Loop
		case token.Whitespace:
			i := strings.LastIndexByte(tok.Text, '\n')
			io.WriteString(w, tok.Text[:i+1])
			ws = tok.Text[i+1:]
		default:
			io.WriteString(w, tok.Text)
		}
	}
	if err := scan.Err(); err != nil {
		var scanErr *token.ScanError
		if errors.As(err, &scanErr) {
			return fmt.Errorf("%s:%v: %v", filename, scanErr.Pos, scanErr.Err)
		}
		return fmt.Errorf("formatting %q: %v", filename, err)
	}
	return w.err
}

type stickyErrWriter struct {
	w   io.Writer
	err error
}

func (w *stickyErrWriter) Write(p []byte) (n int, err error) {
	if w.err != nil {
		return 0, w.err
	}
	n, w.err = w.w.Write(p)
	return n, w.err
}

type Pos struct {
	Line, Column int
}

func (p Pos) Add(q Pos) Pos {
	if q.Line == 1 {
		p.Column += q.Column - 1
	} else {
		p.Line += q.Line - 1
		p.Column = q.Column
	}
	return p
}

func (p Pos) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}
