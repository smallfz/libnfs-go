package fs

import (
	"testing"
)

func TestAbs(t *testing.T) {
	grantedCases := [][]string{
		{"", "/"},
		{".", "/"},
		{"/", "/"},
		{"a/bc/def", "/a/bc/def"},
		{"/a/bc", "/a/bc"},
		{"abc", "/abc"},
		{"./abc", "/abc"},
		{"../abc", "/abc"},
		{"/../abc", "/abc"},
		{"./../abc", "/abc"},
		{"./../abc/def", "/abc/def"},
	}

	for _, row := range grantedCases {
		input := row[0]
		output := Abs(input)
		if output != row[1] {
			t.Fatalf("expects `%s` but get `%s`.", row[1], output)
		}
	}
}

func TestPathJoin(t *testing.T) {
	grantedCases := [][]string{
		{"", "/", "/"},
		{"abc", "/def", "/abc/def"},
		{"/abc", "/def", "/abc/def"},
		{"/abc", "def", "/abc/def"},
		{"/abc", "def", "/ijk", "/abc/def/ijk"},
	}

	for _, row := range grantedCases {
		if len(row) < 3 {
			continue
		}
		output := Abs(Join(row[:len(row)-1]...))
		expect := row[len(row)-1]
		if output != expect {
			t.Fatalf("expects `%s` but get `%s`.", expect, output)
		}
	}
}
