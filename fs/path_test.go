package fs

import (
	"testing"
)

func TestAbs(t *testing.T) {
	grantedCases := [][]string{
		[]string{"", "/"},
		[]string{".", "/"},
		[]string{"/", "/"},
		[]string{"a/bc/def", "/a/bc/def"},
		[]string{"/a/bc", "/a/bc"},
		[]string{"abc", "/abc"},
		[]string{"./abc", "/abc"},
		[]string{"../abc", "/abc"},
		[]string{"/../abc", "/abc"},
		[]string{"./../abc", "/abc"},
		[]string{"./../abc/def", "/abc/def"},
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
		[]string{"", "/", "/"},
		[]string{"abc", "/def", "/abc/def"},
		[]string{"/abc", "/def", "/abc/def"},
		[]string{"/abc", "def", "/abc/def"},
		[]string{"/abc", "def", "/ijk", "/abc/def/ijk"},
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
