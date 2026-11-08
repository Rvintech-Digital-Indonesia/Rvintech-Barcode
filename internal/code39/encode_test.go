package code39

import "testing"

func TestAlphabetPatterns(t *testing.T) {
	patterns := code39Patterns()
	if len(patterns) != 44 {
		t.Fatalf("got %d patterns", len(patterns))
	}
	for r, p := range patterns {
		if len(p) != 9 {
			t.Fatalf("%q: %d elements", r, len(p))
		}
		wide := 0
		for _, c := range p {
			if c == 'w' {
				wide++
			}
		}
		if wide != 3 {
			t.Fatalf("%q: %d wide elements", r, wide)
		}
	}
}
