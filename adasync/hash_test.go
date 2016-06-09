package adasync

import (
	"testing"
)

func TestHashFromBytes(t *testing.T) {
	hash := HashFromBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
	if hash.String() != "AQIDBAUGBwgJCgsMDQ4PEA==" {
		t.Error("Incorrect hash string")
	}
}

func TestEqual(t *testing.T) {
	tests := []struct {
		a, b   *Hash
		expect bool
	}{
		{
			a:      HashFromBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}),
			b:      HashFromBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}),
			expect: true,
		}, {
			a:      nil,
			b:      HashFromBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}),
			expect: false,
		}, {
			a:      nil,
			b:      nil,
			expect: true,
		}, {
			a:      HashFromBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}),
			b:      nil,
			expect: false,
		},
	}
	for _, test := range tests {
		if test.expect != test.a.Equal(test.b) {
			t.Error("Hash equlaity check failed")
		}
	}
}
