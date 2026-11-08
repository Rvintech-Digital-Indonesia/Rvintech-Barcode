package qr

import (
	"strings"
	"testing"
)

func TestAutomaticSegments(t *testing.T) {
	cases := []struct {
		text  string
		modes []int
	}{{"1234567890", []int{numericMode}}, {"HELLO WORLD", []int{alphaMode}}, {"hello", []int{byteMode}}, {"12345678901234567890abcdefghijABCDEFGHIJKLMNOPQRSTUVWXYZ", []int{numericMode, byteMode, alphaMode}}}
	for _, tc := range cases {
		segs, bits := optimizeSegments([]byte(tc.text), 1)
		if bits == 0 {
			t.Fatal("zero cost")
		}
		got := map[int]bool{}
		for _, s := range segs {
			got[s.mode] = true
		}
		for _, mode := range tc.modes {
			if !got[mode] {
				t.Errorf("%q missing mode %d: %#v", tc.text, mode, segs)
			}
		}
	}
}

func TestOptimalSegmentationMatchesExhaustiveCosts(t *testing.T) {
	// Small strings exhaustively cover transitions among all three modes.
	inputs := []string{"1aA", "1234567890ABCxyz", "HELLO world 123", "%%A9abc", "", "0000000000000000"}
	for _, text := range inputs {
		for _, version := range []int{1, 10, 27} {
			got, bits := optimizeSegments([]byte(text), version)
			want := bruteSegmentCost([]byte(text), version)
			if bits != want {
				t.Errorf("%q version %d: got %d bits, want %d", text, version, bits, want)
			}
			decoded := make([]byte, 0, len(text))
			for _, s := range got {
				decoded = append(decoded, s.data...)
			}
			if string(decoded) != text {
				t.Errorf("%q version %d: reconstructed %q", text, version, decoded)
			}
		}
	}
}

func bruteSegmentCost(data []byte, version int) int {
	if len(data) == 0 {
		return 0
	}
	const inf = int(^uint(0)>>1) / 4
	cost := make([]int, len(data)+1)
	for i := 1; i < len(cost); i++ {
		cost[i] = inf
	}
	for start := 0; start < len(data); start++ {
		for _, mode := range []int{numericMode, alphaMode, byteMode} {
			for end := start + 1; end <= len(data) && end-start < (1<<countWidth(mode, version)); end++ {
				valid := true
				for _, b := range data[start:end] {
					if mode == numericMode && (b < '0' || b > '9') || mode == alphaMode && alphaIndex(b) < 0 {
						valid = false
						break
					}
				}
				if !valid {
					break
				}
				candidate := cost[start] + 4 + countWidth(mode, version) + payloadBits(mode, end-start)
				if candidate < cost[end] {
					cost[end] = candidate
				}
			}
		}
	}
	return cost[len(data)]
}

func TestMaximumBytePayloadSegmentation(t *testing.T) {
	for level := 0; level < 4; level++ {
		capacityBits := dataCodewords(40, level) * 8
		maxBytes := (capacityBits - 20) / 8
		text := make([]byte, maxBytes)
		for i := range text {
			text[i] = 'a'
		}
		segs, bits := optimizeSegments(text, 40)
		if bits != 20+len(text)*8 || bits > capacityBits {
			t.Errorf("level %d maximum byte payload uses %d bits, capacity %d", level, bits, capacityBits)
		}
		if len(segs) != 1 || segs[0].mode != byteMode || len(segs[0].data) != len(text) {
			t.Errorf("level %d unexpected maximum payload segmentation: %d segments", level, len(segs))
		}
		_, overCapacityBits := optimizeSegments(append(text, 'a'), 40)
		if overCapacityBits <= capacityBits {
			t.Errorf("level %d over-capacity byte payload costs %d bits, capacity %d", level, overCapacityBits, capacityBits)
		}
	}
}

func TestHelloWorldVersion1QReferenceVector(t *testing.T) {
	// Worked Version 1-Q vector: https://useqr.app/docs/spec/encode-hello-world-by-hand
	segs, bits := optimizeSegments([]byte("HELLO WORLD"), 1)
	if bits != 74 || len(segs) != 1 || segs[0].mode != alphaMode {
		t.Fatalf("unexpected HELLO WORLD segmentation: %d bits, %#v", bits, segs)
	}
	data := encodeSegments(segs, dataCodewords(1, 2), 1)
	wantData := []byte{0x20, 0x5B, 0x0B, 0x78, 0xD1, 0x72, 0xDC, 0x4D, 0x43, 0x40, 0xEC, 0x11, 0xEC}
	if string(data) != string(wantData) {
		t.Fatalf("data codewords % X, want % X", data, wantData)
	}
	wantECC := []byte{0xA8, 0x48, 0x16, 0x52, 0xD9, 0x36, 0x9C, 0x00, 0x2E, 0x0F, 0xB4, 0x7A, 0x10}
	gotECC := reedSolomon(data, len(wantECC))
	if string(gotECC) != string(wantECC) {
		t.Fatalf("ECC codewords % X, want % X", gotECC, wantECC)
	}
}

func BenchmarkOptimizeSegments(b *testing.B) {
	for _, tc := range []struct {
		name string
		text []byte
	}{
		{"short", []byte("hello 123")},
		{"medium", []byte("invoice-1234567890-invoice-1234567890-invoice-1234567890-invoice-1234567890")},
		{"maximum", []byte(strings.Repeat("a", 2953))},
	} {
		b.Run(tc.name, func(b *testing.B) {
			b.SetBytes(int64(len(tc.text)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				optimizeSegments(tc.text, 40)
			}
		})
	}
}

func TestAutomaticDoesNotExceedByteEncoding(t *testing.T) {
	for _, text := range []string{"1234567890", "HELLO WORLD", "123abcXYZ", "日本語 QR 123"} {
		segs, bits := optimizeSegments([]byte(text), 1)
		byteBits := 4 + 8 + len([]byte(text))*8
		if bits > byteBits {
			t.Errorf("%q optimized bits %d > byte bits %d", text, bits, byteBits)
		}
		got := 0
		for _, s := range segs {
			got += len(s.data)
		}
		if got != len([]byte(text)) {
			t.Errorf("%q lost bytes", text)
		}
	}
}

func TestFunctionModuleCounts(t *testing.T) {
	for v := 1; v <= 40; v++ {
		_, fn := qrBase(v)
		free := 0
		for _, row := range fn {
			for _, reserved := range row {
				if !reserved {
					free++
				}
			}
		}
		if free != rawDataModules(v) {
			t.Fatalf("version %d: data modules %d, want %d", v, free, rawDataModules(v))
		}
	}
}
func TestCodewordInterleavingDimensions(t *testing.T) {
	for v := 1; v <= 40; v++ {
		for level := 0; level < 4; level++ {
			cap := dataCodewords(v, level)
			data := make([]byte, cap)
			out := interleave(data, v, level)
			if len(out) != rawDataModules(v)/8 {
				t.Fatalf("version %d level %d: %d codewords, want %d", v, level, len(out), rawDataModules(v)/8)
			}
		}
	}
}
