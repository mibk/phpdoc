// PHPDocfmt formats PHPDoc comments in PHP files.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"unicode/utf8"

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
			log.Fatalf("formatting %s: %v", filename, err)
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

var phpdocRx = regexp.MustCompile(`(?s)[ \t]*/\*\*.*?\*/[ \t]*\n?`)

func formatDocs(filename string, out io.Writer, in io.Reader) error {
	data, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}

	pos := Position{Filename: filename, Line: 1, Column: 1}
	w := &stickyErrWriter{w: out}
	for len(data) > 0 {
		loc := phpdocRx.FindIndex(data)
		if loc == nil {
			w.Write(data)
			break
		}

		m, n := loc[0], loc[1]
		if m > 0 {
			pos.Shift(data[:m])
			w.Write(data[:m])
		}

		if doc, err := phpdoc.Parse(bytes.NewReader(data[m:n])); err != nil {
			errPos := pos
			if se, ok := err.(*phpdoc.SyntaxError); ok {
				errPos = errPos.Add(Position{Line: se.Line, Column: se.Column})
				err = se.Err
			}
			log.Printf("%s: %v", errPos, err)
			w.Write(data[m:n])
		} else {
			if err := phpdoc.Fprint(w, doc); err != nil {
				log.Printf("%s: printing doc: %v", pos, err)
			}
		}
		pos.Shift(data[m:n])
		data = data[n:]
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

type Position struct {
	Filename     string
	Line, Column int
}

func (p Position) Add(q Position) Position {
	if q.Line == 1 {
		p.Column += q.Column - 1
	} else {
		p.Line += q.Line - 1
		p.Column = q.Column
	}
	return p
}

func (p *Position) Shift(b []byte) {
	q := endPosition(b)
	*p = p.Add(q)
}

func endPosition(b []byte) Position {
	lines := bytes.Count(b, []byte("\n"))
	i := bytes.LastIndexByte(b, '\n')
	columns := utf8.RuneCount(b[i+1:])
	return Position{Line: lines + 1, Column: columns + 1}
}

func (p Position) String() string {
	return fmt.Sprintf("%s:%d:%d", p.Filename, p.Line, p.Column)
}
