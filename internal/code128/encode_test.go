package code128

import "testing"

func TestPatternWidths(t *testing.T) {
	patterns := code128Patterns()
	if len(patterns) != 107 {
		t.Fatalf("got %d patterns", len(patterns))
	}
	for i, p := range patterns {
		sum := 0
		for _, w := range p {
			sum += int(w - '0')
		}
		want := 11
		if i == 106 {
			want = 13
		}
		if sum != want {
			t.Fatalf("codeword %d width=%d", i, sum)
		}
	}
}

func TestCodeSetAControlMapping(t *testing.T) {
	got := code128Values([]byte{0})
	want := []int{103, 64, 64, 106}
	for i, v := range want {
		if got[i] != v {
			t.Fatalf("codeword %d = %d, want %d", i, got[i], v)
		}
	}
}
