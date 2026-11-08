package code128

import (
	"fmt"
	"github.com/Rvintech-Digital-Indonesia/Rvintech-Barcode/internal/symbol"
	"strings"
)

// EncodeCode128 encodes ASCII input, switching automatically between Code Sets
// A, B, and C. Code C is selected for digit runs; A handles control characters.
func Encode(text string, options symbol.Options) (*symbol.Code, error) {
	if text == "" {
		return nil, symbol.ErrEmptyInput
	}
	if len(text) > 512 {
		return nil, fmt.Errorf("barcode: Code 128 input exceeds 512 bytes")
	}
	o, err := symbol.NormalizeOptions(options)
	if err != nil {
		return nil, err
	}
	if options.QuietZone == 0 {
		o.QuietZone = 10
	}
	if o.QuietZone < 10 {
		return nil, fmt.Errorf("barcode: Code 128 quiet zone must be at least ten modules")
	}
	if o.Logo != nil || o.Shape != symbol.ShapeSquare {
		return nil, fmt.Errorf("barcode: Code 128 supports colors but not logos or QR module shapes")
	}
	raw := []byte(text)
	for _, b := range raw {
		if b > 127 {
			return nil, fmt.Errorf("barcode: Code 128 currently supports ASCII input only")
		}
	}
	codes := code128Values(raw)
	patterns := code128Patterns()
	bits := []bool{}
	for _, v := range codes {
		p := patterns[v]
		for i := 0; i < len(p); i++ {
			width := int(p[i] - '0')
			on := i%2 == 0
			for k := 0; k < width; k++ {
				bits = append(bits, on)
			}
		}
	}
	return symbol.New(linearMatrix(bits), o), nil
}

func code128Values(raw []byte) []int {
	codes := []int{}
	pos := 0
	set := byte('B')
	if digitRun(raw, 0) >= 4 {
		set = 'C'
	} else if raw[0] < 32 {
		set = 'A'
	}
	switch set {
	case 'A':
		codes = append(codes, 103)
	case 'B':
		codes = append(codes, 104)
	case 'C':
		codes = append(codes, 105)
	}
	for pos < len(raw) {
		if set != 'C' && digitRun(raw, pos) >= 4 {
			codes = append(codes, 99)
			set = 'C'
			continue
		}
		if set == 'C' {
			if pos+1 < len(raw) && isDigit(raw[pos]) && isDigit(raw[pos+1]) {
				codes = append(codes, int(raw[pos]-'0')*10+int(raw[pos+1]-'0'))
				pos += 2
				continue
			}
			if raw[pos] < 32 {
				codes = append(codes, 101)
				set = 'A'
			} else {
				codes = append(codes, 100)
				set = 'B'
			}
			continue
		}
		if set == 'B' && raw[pos] < 32 {
			codes = append(codes, 101)
			set = 'A'
			continue
		}
		if set == 'A' && raw[pos] >= 96 {
			codes = append(codes, 100)
			set = 'B'
			continue
		}
		if set == 'A' {
			value := int(raw[pos]) - 32
			if raw[pos] < 32 {
				value = int(raw[pos]) + 64
			}
			codes = append(codes, value)
		} else {
			codes = append(codes, int(raw[pos]-32))
		}
		pos++
	}
	sum := codes[0]
	for i := 1; i < len(codes); i++ {
		sum += codes[i] * i
	}
	codes = append(codes, sum%103, 106)
	return codes
}

func isDigit(b byte) bool { return b >= '0' && b <= '9' }
func digitRun(b []byte, pos int) int {
	n := 0
	for pos < len(b) && isDigit(b[pos]) {
		n++
		pos++
	}
	return n
}
func code128Patterns() []string {
	return strings.Fields(`212222 222122 222221 121223 121322 131222 122213 122312 132212 221213 221312 231212 112232 122132 122231 113222 123122 123221 223211 221132 221231 213212 223112 312131 311222 321122 321221 312212 322112 322211 212123 212321 232121 111323 131123 131321 112313 132113 132311 211313 231113 231311 112133 112331 132131 113123 113321 133121 313121 211331 231131 213113 213311 213131 311123 311321 331121 312113 312311 332111 314111 221411 431111 111224 111422 121124 121421 141122 141221 112214 112412 122114 122411 142112 142211 241211 221114 413111 241112 134111 111242 121142 121241 114212 124112 124211 411212 421112 421211 212141 214121 412121 111143 111341 131141 114113 114311 411113 411311 113141 114131 311141 411131 211412 211214 211232 2331112`)
}

func linearMatrix(bits []bool) [][]bool {
	rows := make([][]bool, 40)
	for y := range rows {
		rows[y] = append([]bool(nil), bits...)
	}
	return rows
}
