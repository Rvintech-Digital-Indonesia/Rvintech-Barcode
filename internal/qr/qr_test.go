package qr

import "testing"

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
