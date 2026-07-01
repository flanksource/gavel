package sample

import (
	"fmt"
	"testing"
)

func TestAdd(t *testing.T) {
	if add(1, 2) != 3 {
		t.Errorf("expected 3")
	}
}

func TestTable(t *testing.T) {
	cases := []struct {
		name string
		in   int
		want int
	}{
		{"zero value", 0, 0},
		{"positive", 2, 4},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if double(tc.in) != tc.want {
				t.Errorf("double(%d) != %d", tc.in, tc.want)
			}
		})
	}
}

func TestSubtests(t *testing.T) {
	t.Run("literal child", func(t *testing.T) {
		t.Run("nested grandchild", func(t *testing.T) {
			if add(1, 1) != 2 && double(1) != 2 {
				t.Fail()
			}
		})
	})
	t.Run(fmt.Sprintf("case-%d", 1), func(t *testing.T) {})
}

func TestMain(m *testing.M) {}

func helperNotATest(t *testing.T, n int) {}

func add(a, b int) int { return a + b }

func double(n int) int { return n * 2 }
