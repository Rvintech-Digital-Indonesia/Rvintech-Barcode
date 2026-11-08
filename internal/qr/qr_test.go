package qr

import "testing"

func TestAutomaticSegments(t *testing.T) {
	cases := []struct{text string; modes []int}{{"1234567890", []int{numericMode}}, {"HELLO WORLD", []int{alphaMode}}, {"hello", []int{byteMode}}, {"123abcXYZ", []int{numericMode, byteMode, alphaMode}}}
	for _, tc := range cases { segs, bits := optimizeSegments([]byte(tc.text),1); if bits==0 {t.Fatal("zero cost")}; got:=map[int]bool{};for _,s:=range segs {got[s.mode]=true};for _,mode:=range tc.modes {if !got[mode] {t.Errorf("%q missing mode %d: %#v",tc.text,mode,segs)}} }
}

func TestAutomaticDoesNotExceedByteEncoding(t *testing.T) {
	for _,text:=range []string{"1234567890","HELLO WORLD","123abcXYZ","日本語 QR 123"} { segs,bits:=optimizeSegments([]byte(text),1); byteBits:=4+8+len([]byte(text))*8; if bits>byteBits {t.Errorf("%q optimized bits %d > byte bits %d",text,bits,byteBits)}; got:=0;for _,s:=range segs {got+=len(s.data)};if got!=len([]byte(text)){t.Errorf("%q lost bytes",text)} }
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
