package main

import (
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"

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
		if err := formatDocs(os.Stdout, os.Stdin); err != nil {
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
		if err := formatDocs(buf, f); err != nil {
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

func formatDocs(out io.Writer, in io.Reader) error {
	data, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}

	w := &stickyErrWriter{w: out}
	for len(data) > 0 {
		loc := phpdocRx.FindIndex(data)
		if loc == nil {
			w.Write(data)
			break
		}

		m, n := loc[0], loc[1]
		if m > 0 {
			w.Write(data[:m])
		}

		p := phpdoc.NewParser(bytes.NewReader(data[m:n]))
		if doc, err := p.Parse(); err != nil {
			log.Println(err)
			w.Write(data[m:n])
		} else {
			// TODO: Check error?
			if err := phpdoc.Fprint(w, doc); err != nil {
				log.Println(err)
			}
		}

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
