package code39

import (
	"fmt"
	"github.com/pt-rvintech/rvintech-barcode/internal/symbol"
	"strings"
)

// EncodeCode39 encodes the Code 39 alphabet. Checksum enables the optional
// modulo-43 check character. Input is case-insensitive and normalized to upper case.
func Encode(text string, options symbol.Options) (*symbol.Code, error) {
	if text == "" {
		return nil, symbol.ErrEmptyInput
	}
	if len(text) > 512 {
		return nil, fmt.Errorf("barcode: Code 39 input exceeds 512 bytes")
	}
	o, err := symbol.NormalizeOptions(options)
	if err != nil {
		return nil, err
	}
	if options.QuietZone == 0 {
		o.QuietZone = 10
	}
	if o.QuietZone < 10 {
		return nil, fmt.Errorf("barcode: Code 39 quiet zone must be at least ten modules")
	}
	if o.Logo != nil || o.Shape != symbol.ShapeSquare {
		return nil, fmt.Errorf("barcode: Code 39 supports colors but not logos or QR module shapes")
	}
	text = strings.ToUpper(text)
	for _, r := range text {
		if r == '*' || strings.IndexRune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ-. $/+%", r) < 0 {
			return nil, fmt.Errorf("barcode: Code 39 does not support character %q", r)
		}
	}
	const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ-. $/+%"
	if o.Checksum {
		sum := 0
		for _, r := range text {
			i := strings.IndexRune(alphabet, r)
			if i < 0 {
				return nil, fmt.Errorf("barcode: Code 39 does not support character %q", r)
			}
			sum += i
		}
		text += string(alphabet[sum%43])
	}
	patterns := code39Patterns()
	bits := make([]bool, 0, len(text)*16)
	for index, r := range "*" + text + "*" {
		p, ok := patterns[byte(r)]
		if !ok {
			return nil, fmt.Errorf("barcode: Code 39 does not support character %q", r)
		}
		for j := 0; j < 9; j++ {
			wide := p[j] == 'w'
			width := 1
			if wide {
				width = 3
			}
			on := j%2 == 0
			for k := 0; k < width; k++ {
				bits = append(bits, on)
			}
		}
		if index < len(text)+1 {
			bits = append(bits, false)
		}
	}
	return symbol.New(linearMatrix(bits), o), nil
}
func code39Patterns() map[byte]string {
	return map[byte]string{
		'0': "nnnwwnwnn", '1': "wnnwnnnnw", '2': "nnwwnnnnw", '3': "wnwwnnnnn", '4': "nnnwwnnnw", '5': "wnnwwnnnn", '6': "nnwwwnnnn", '7': "nnnwnnwnw", '8': "wnnwnnwnn", '9': "nnwwnnwnn",
		'A': "wnnnnwnnw", 'B': "nnwnnwnnw", 'C': "wnwnnwnnn", 'D': "nnnnwwnnw", 'E': "wnnnwwnnn", 'F': "nnwnwwnnn", 'G': "nnnnnwwnw", 'H': "wnnnnwwnn", 'I': "nnwnnwwnn", 'J': "nnnnwwwnn",
		'K': "wnnnnnnww", 'L': "nnwnnnnww", 'M': "wnwnnnnwn", 'N': "nnnnwnnww", 'O': "wnnnwnnwn", 'P': "nnwnwnnwn", 'Q': "nnnnnnwww", 'R': "wnnnnnwwn", 'S': "nnwnnnwwn", 'T': "nnnnwnwwn",
		'U': "wwnnnnnnw", 'V': "nwwnnnnnw", 'W': "wwwnnnnnn", 'X': "nwnnwnnnw", 'Y': "wwnnwnnnn", 'Z': "nwwnwnnnn", '-': "nwnnnnwnw", '.': "wwnnnnwnn", ' ': "nwwnnnwnn", '$': "nwnwnwnnn", '/': "nwnwnnnwn", '+': "nwnnnwnwn", '%': "nnnwnwnwn", '*': "nwnnwnwnn",
	}
}

func linearMatrix(bits []bool) [][]bool {
	rows := make([][]bool, 40)
	for y := range rows {
		rows[y] = append([]bool(nil), bits...)
	}
	return rows
}
