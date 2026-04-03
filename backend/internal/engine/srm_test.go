//go:build !integration

package engine

import "testing"

func TestComputeSRMFallback(t *testing.T) {
	cases := []struct {
		c    string
		h    float64
		want string
	}{
		{"L3", 73, "L1"},
		{"L3", 24, "L3"},
		{"L2", 50, "L2"},
	}
	for _, tc := range cases {
		if got := ComputeSRMFallback(tc.c, tc.h); got != tc.want {
			t.Errorf("got %q want %q", got, tc.want)
		}
	}
}
