package syntax

import (
	"fmt"
	"testing"
	"unicode"
)

// foldPerRune is the reference: close a set under case folding by visiting
// every rune in every range.
func foldPerRune(ranges []SingleRange, options RegexOptions) CharSet {
	c := CharSet{ranges: append([]SingleRange(nil), ranges...)}
	rangeCount := len(c.ranges)
	for i := 0; i < rangeCount; i++ {
		r := c.ranges[i]
		for ch := r.First; ch <= r.Last; ch++ {
			for _, eq := range tryFindCaseEquivalences(ch, options) {
				c.ranges = append(c.ranges, SingleRange{First: eq, Last: eq})
			}
		}
	}
	c.canonicalize()
	return c
}

func foldOverFoldableRunes(ranges []SingleRange, options RegexOptions) CharSet {
	c := CharSet{ranges: append([]SingleRange(nil), ranges...)}
	c.addCaseEquivalences(options)
	return c
}

// TestAddCaseEquivalencesMatchesPerRuneFolding pins the result to what folding
// one rune at a time produces, since that is the behaviour patterns already
// depend on.
func TestAddCaseEquivalencesMatchesPerRuneFolding(t *testing.T) {
	sets := map[string][]SingleRange{
		"lower ascii":     {{First: 'a', Last: 'z'}},
		"negated lower":   {{First: 0, Last: 'a' - 1}, {First: 'z' + 1, Last: unicode.MaxRune}},
		"upper ascii":     {{First: 'A', Last: 'Z'}},
		"greek":           {{First: 0x03B1, Last: 0x03C9}},
		"latin extended":  {{First: 0x00C0, Last: 0x024F}},
		"kelvin sign":     {{First: 0x212A, Last: 0x212A}},
		"sharp s":         {{First: 0x00DF, Last: 0x00DF}},
		"dotless i":       {{First: 0x0131, Last: 0x0131}},
		"everything":      {{First: 0, Last: unicode.MaxRune}},
		"disjoint points": {{First: 'a', Last: 'a'}, {First: 0x1E9E, Last: 0x1E9E}, {First: 0x10400, Last: 0x10427}},
	}

	options := map[string]RegexOptions{
		"none":                 0,
		"ignore case":          IgnoreCase,
		"ecmascript":           ECMAScript | IgnoreCase,
		"ecmascript unicode":   ECMAScript | Unicode | IgnoreCase,
		"unicode":              Unicode | IgnoreCase,
		"multiline ignorecase": Multiline | IgnoreCase,
	}

	for setName, ranges := range sets {
		for optionName, opts := range options {
			name := fmt.Sprintf("%s/%s", setName, optionName)
			t.Run(name, func(t *testing.T) {
				want := foldPerRune(ranges, opts)
				got := foldOverFoldableRunes(ranges, opts)

				if len(got.ranges) != len(want.ranges) {
					t.Fatalf("ranges: got %d, want %d", len(got.ranges), len(want.ranges))
				}
				for i := range want.ranges {
					if got.ranges[i] != want.ranges[i] {
						t.Fatalf("range %d: got %v, want %v", i, got.ranges[i], want.ranges[i])
					}
				}
			})
		}
	}
}

// BenchmarkAddCaseEquivalences measures closing a set under folding, against
// the per-rune reference. A negated class under ignore case covers most of
// Unicode, which is what made this worth changing.
func BenchmarkAddCaseEquivalences(b *testing.B) {
	sets := map[string][]SingleRange{
		"lower ascii":   {{First: 'a', Last: 'z'}},
		"negated lower": {{First: 0, Last: 'a' - 1}, {First: 'z' + 1, Last: unicode.MaxRune}},
	}

	for name, ranges := range sets {
		b.Run(name+"/per-rune", func(b *testing.B) {
			for b.Loop() {
				foldPerRune(ranges, ECMAScript|IgnoreCase)
			}
		})
		b.Run(name+"/foldable-runes", func(b *testing.B) {
			for b.Loop() {
				foldOverFoldableRunes(ranges, ECMAScript|IgnoreCase)
			}
		})
	}
}
