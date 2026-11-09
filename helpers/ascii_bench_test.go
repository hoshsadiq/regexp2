package helpers

import (
	"strings"
	"testing"
)

// BenchmarkDecodeString_Mixed covers input that misses the all-ASCII fast path.
func BenchmarkDecodeString_Mixed(b *testing.B) {
	s := strings.Repeat("nümber=+71112223334 SSN:123-45-6789 Й ", 1<<12)
	buf := make([]rune, len(s))
	b.SetBytes(int64(len(s)))
	b.ReportAllocs()
	for b.Loop() {
		if n, ascii := DecodeString(s, buf); ascii || n == 0 {
			b.Fatal(n, ascii)
		}
	}
}
