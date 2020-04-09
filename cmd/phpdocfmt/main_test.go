package main

import "testing"

func pos(line, col int) Position {
	return Position{Line: line, Column: col}
}

func TestPosition_endPosition(t *testing.T) {
	tests := []struct {
		text string
		want Position
	}{
		{"", pos(1, 1)},
		{"x", pos(1, 2)},
		{"x\n", pos(2, 1)},
		{"12345678", pos(1, 9)},
		{"123456789\n", pos(2, 1)},
		{"123456789\n123", pos(2, 4)},
		{"123456789\n123\n", pos(3, 1)},
	}

	for _, tt := range tests {
		if got := endPosition([]byte(tt.text)); got != tt.want {
			t.Errorf("%q: got %v, want %v", tt.text, got, tt.want)
		}
	}
}

func TestPostion_Add(t *testing.T) {
	tests := []struct {
		p, q Position
		want Position
	}{
		{pos(1, 1), pos(1, 1), pos(1, 1)},
		{pos(2, 2), pos(1, 1), pos(2, 2)},
		{pos(100, 30), pos(1, 1), pos(100, 30)},
		{pos(100, 30), pos(1, 2), pos(100, 31)},
		{pos(100, 30), pos(1, 3), pos(100, 32)},
		{pos(100, 30), pos(1, 71), pos(100, 100)},
		{pos(10, 20), pos(2, 3), pos(11, 3)},
		{pos(10, 20), pos(2, 4), pos(11, 4)},
		{pos(10, 20), pos(2, 48), pos(11, 48)},
		{pos(10, 20), pos(12, 48), pos(21, 48)},
	}

	for _, tt := range tests {
		if got := tt.p.Add(tt.q); got != tt.want {
			t.Errorf("%v + %v = %v; want %v", tt.p, tt.q, got, tt.want)
		}
	}
}
