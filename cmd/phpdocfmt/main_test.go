package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPostion_Add(t *testing.T) {
	tests := []struct {
		p, q Pos
		want Pos
	}{
		{Pos{1, 1}, Pos{1, 1}, Pos{1, 1}},
		{Pos{2, 2}, Pos{1, 1}, Pos{2, 2}},
		{Pos{100, 30}, Pos{1, 1}, Pos{100, 30}},
		{Pos{100, 30}, Pos{1, 2}, Pos{100, 31}},
		{Pos{100, 30}, Pos{1, 3}, Pos{100, 32}},
		{Pos{100, 30}, Pos{1, 71}, Pos{100, 100}},
		{Pos{10, 20}, Pos{2, 3}, Pos{11, 3}},
		{Pos{10, 20}, Pos{2, 4}, Pos{11, 4}},
		{Pos{10, 20}, Pos{2, 48}, Pos{11, 48}},
		{Pos{10, 20}, Pos{12, 48}, Pos{21, 48}},
	}

	for _, tt := range tests {
		if got := tt.p.Add(tt.q); got != tt.want {
			t.Errorf("%v + %v = %v; want %v", tt.p, tt.q, got, tt.want)
		}
	}
}

func TestFormatting(t *testing.T) {
	files, err := filepath.Glob("testdata/*.input")
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".input")
		t.Run(name, func(t *testing.T) {
			f, err := os.Open(file)
			if err != nil {
				t.Fatal(err)
			}
			buf := new(bytes.Buffer)
			if err := formatDocs(name, buf, f); err != nil {
				t.Fatal(err)
			}
			f.Close()

			want, err := ioutil.ReadFile(filepath.Join("testdata", name+".golden"))
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(strings.Split(buf.String(), "\n"), strings.Split(string(want), "\n")); diff != "" {
				t.Errorf("files don't match (-got +want)\n%s", diff)
			}
		})
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{{
		"unterminated",
		`<?php /**`,
		"unterminated:1:10: unterminated block comment",
	}, {
		"invalid.method.tag",
		`<?php /**  @method */`,
		"invalid.method.tag:1:20: expecting ( or basic type, found */",
	}}

	for _, tt := range tests {
		errStr := "<nil>"
		src := strings.NewReader(tt.input)
		if err := formatDocs(tt.name, ioutil.Discard, src); err != nil {
			errStr = err.Error()
		}
		if errStr != tt.wantErr {
			t.Errorf("%s:\n got %s\nwant %s", tt.name, errStr, tt.wantErr)
		}
	}
}
