package qr

import (
	"fmt"
	"github.com/Rvintech-Digital-Indonesia/Rvintech-Barcode/internal/symbol"
)

// Encode creates a QR Code Model 2 symbol in byte mode, versions 1 through 40.
func Encode(text string, options symbol.Options) (*symbol.Code, error) {
	if text == "" {
		return nil, symbol.ErrEmptyInput
	}
	o, err := symbol.NormalizeOptions(options)
	if err != nil {
		return nil, err
	}
	if o.QuietZone < 4 {
		return nil, fmt.Errorf("barcode: QR quiet zone must be at least four modules")
	}
	level := 0
	switch o.ErrorCorrection {
	case symbol.ErrorCorrectionM:
		level = 1
	case symbol.ErrorCorrectionQ:
		level = 2
	case symbol.ErrorCorrectionH:
		level = 3
	}
	if o.Logo != nil && level < 3 {
		level = 3
		o.ErrorCorrection = symbol.ErrorCorrectionH
	}
	data := []byte(text)
	version := 0
	for v := 1; v <= 40; v++ {
		countBits := 8
		if v >= 10 {
			countBits = 16
		}
		if len(data) < 1<<countBits && 4+countBits+len(data)*8 <= dataCodewords(v, level)*8 {
			if o.Logo != nil && logoTouchesFunction(v, o.QuietZone) {
				continue
			}
			version = v
			break
		}
	}
	if version == 0 {
		return nil, fmt.Errorf("barcode: QR payload exceeds version 40 capacity at error correction level %s", o.ErrorCorrection)
	}
	capacity := dataCodewords(version, level)
	all := interleave(qrDataCodewords(data, capacity, version), version, level)
	base, fn := qrBase(version)
	bestScore := int(^uint(0) >> 1)
	var best [][]bool
	for mask := 0; mask < 8; mask++ {
		m := cloneMatrix(base)
		placeData(m, fn, all, mask)
		placeFormat(m, mask, level)
		score := qrPenalty(m)
		if score < bestScore {
			best, bestScore = m, score
		}
	}
	return symbol.NewProtected(best, fn, o), nil
}

func qrDataCodewords(data []byte, capacity int, version int) []byte {
	bits := make([]bool, 0, capacity*8)
	putBits := func(v uint, n int) {
		for i := n - 1; i >= 0; i-- {
			bits = append(bits, (v>>i)&1 != 0)
		}
	}
	putBits(4, 4)
	if version < 10 {
		putBits(uint(len(data)), 8)
	} else {
		putBits(uint(len(data)), 16)
	}
	for _, b := range data {
		putBits(uint(b), 8)
	}
	for i := 0; i < 4 && len(bits) < capacity*8; i++ {
		bits = append(bits, false)
	}
	for len(bits)%8 != 0 {
		bits = append(bits, false)
	}
	out := make([]byte, 0, capacity)
	for i := 0; i < len(bits); i += 8 {
		var v byte
		for j := 0; j < 8; j++ {
			v <<= 1
			if bits[i+j] {
				v |= 1
			}
		}
		out = append(out, v)
	}
	pad := byte(0xEC)
	for len(out) < capacity {
		out = append(out, pad)
		pad ^= 0xFD
	}
	return out
}
func reedSolomon(data []byte, degree int) []byte {
	gen := []byte{1}
	for i := 0; i < degree; i++ {
		next := make([]byte, len(gen)+1)
		for j, v := range gen {
			next[j] ^= v
			next[j+1] ^= gfMul(v, gfExp(i))
		}
		gen = next
	}
	rem := make([]byte, degree)
	for _, b := range data {
		factor := b ^ rem[0]
		copy(rem, rem[1:])
		rem[degree-1] = 0
		for j := range rem {
			rem[j] ^= gfMul(gen[j+1], factor)
		}
	}
	return rem
}
func gfExp(n int) byte {
	x := byte(1)
	for i := 0; i < n; i++ {
		x = gfMul(x, 2)
	}
	return x
}
func gfMul(x, y byte) byte {
	z := byte(0)
	for y != 0 {
		if y&1 != 0 {
			z ^= x
		}
		y >>= 1
		hi := x & 0x80
		x <<= 1
		if hi != 0 {
			x ^= 0x1D
		}
	}
	return z
}
func qrBase(version int) ([][]bool, [][]bool) {
	n := 17 + 4*version
	m := make([][]bool, n)
	fn := make([][]bool, n)
	for y := range m {
		m[y] = make([]bool, n)
		fn[y] = make([]bool, n)
	}
	set := func(x, y int, v bool) {
		if x >= 0 && x < n && y >= 0 && y < n {
			m[y][x] = v
			fn[y][x] = true
		}
	}
	finder := func(cx, cy int) {
		for dy := -4; dy <= 4; dy++ {
			for dx := -4; dx <= 4; dx++ {
				x, y := cx+dx, cy+dy
				if x < 0 || x >= n || y < 0 || y >= n {
					continue
				}
				d := maxIntQR(absInt(dx), absInt(dy))
				set(x, y, d != 2 && d != 4)
			}
		}
	}
	finder(3, 3)
	finder(n-4, 3)
	finder(3, n-4)
	for i := 8; i < n-8; i++ {
		set(6, i, i%2 == 0)
		set(i, 6, i%2 == 0)
	}
	if version >= 2 {
		count := version/7 + 2
		step := (version*8 + count*3 + 5) / (count*4 - 4) * 2
		positions := make([]int, count)
		positions[0] = 6
		for i, pos := count-1, n-7; i >= 1; i, pos = i-1, pos-step {
			positions[i] = pos
		}
		for i, x := range positions {
			for j, y := range positions {
				if (i == 0 && j == 0) || (i == 0 && j == count-1) || (i == count-1 && j == 0) {
					continue
				}
				for dy := -2; dy <= 2; dy++ {
					for dx := -2; dx <= 2; dx++ {
						set(x+dx, y+dy, maxIntQR(absInt(dx), absInt(dy)) != 1)
					}
				}
			}
		}
	}
	if version >= 7 {
		rem := version
		for i := 0; i < 12; i++ {
			rem = (rem << 1) ^ ((rem >> 11) * 0x1F25)
		}
		bits := version<<12 | rem
		for i := 0; i < 18; i++ {
			v := (bits>>i)&1 != 0
			x := n - 11 + i%3
			y := i / 3
			set(x, y, v)
			set(y, x, v)
		}
	}
	// Reserve format-information cells and the fixed dark module.
	for i := 0; i < 9; i++ {
		if i != 6 {
			set(8, i, false)
			set(i, 8, false)
		}
	}
	for i := 0; i < 8; i++ {
		set(n-1-i, 8, false)
		set(8, n-1-i, false)
	}
	set(8, n-8, true)
	return m, fn
}
func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
func cloneMatrix(a [][]bool) [][]bool {
	b := make([][]bool, len(a))
	for i := range a {
		b[i] = append([]bool(nil), a[i]...)
	}
	return b
}
func placeData(m, fn [][]bool, data []byte, mask int) {
	n := len(m)
	bit := 0
	up := true
	for right := n - 1; right >= 1; right -= 2 {
		if right == 6 {
			right--
		}
		for vert := 0; vert < n; vert++ {
			y := vert
			if up {
				y = n - 1 - vert
			}
			for j := 0; j < 2; j++ {
				x := right - j
				if fn[y][x] {
					continue
				}
				v := bit < len(data)*8 && ((data[bit/8]>>(7-uint(bit%8)))&1) != 0
				bit++
				if maskBit(mask, x, y) {
					v = !v
				}
				m[y][x] = v
			}
		}
		up = !up
	}
}
func maskBit(mask, x, y int) bool {
	switch mask {
	case 0:
		return (x+y)%2 == 0
	case 1:
		return y%2 == 0
	case 2:
		return x%3 == 0
	case 3:
		return (x+y)%3 == 0
	case 4:
		return (y/2+x/3)%2 == 0
	case 5:
		return x*y%2+x*y%3 == 0
	case 6:
		return (x*y%2+x*y%3)%2 == 0
	default:
		return ((x+y)%2+x*y%3)%2 == 0
	}
}
func placeFormat(m [][]bool, mask int, level int) {
	n := len(m)
	formatLevel := []int{1, 0, 3, 2}[level]
	data := formatLevel<<3 | mask
	rem := data
	for i := 0; i < 10; i++ {
		rem = (rem << 1) ^ ((rem >> 9) * 0x537)
	}
	bits := ((data << 10) | rem) ^ 0x5412
	bit := func(i int) bool { return (bits>>i)&1 != 0 }
	for i := 0; i <= 5; i++ {
		m[i][8] = bit(i)
	}
	m[7][8] = bit(6)
	m[8][8] = bit(7)
	m[8][7] = bit(8)
	for i := 9; i < 15; i++ {
		m[8][14-i] = bit(i)
	}
	for i := 0; i < 8; i++ {
		m[8][n-1-i] = bit(i)
	}
	for i := 8; i < 15; i++ {
		m[n-15+i][8] = bit(i)
	}
	m[n-8][8] = true
}
func qrPenalty(m [][]bool) int {
	n := len(m)
	score := 0
	for axis := 0; axis < 2; axis++ {
		for a := 0; a < n; a++ {
			run, prev := 0, false
			line := make([]bool, n)
			for b := 0; b < n; b++ {
				x, y := a, b
				if axis == 1 {
					x, y = b, a
				}
				line[b] = m[y][x]
				if b == 0 || line[b] != prev {
					if run >= 5 {
						score += 3 + run - 5
					}
					run = 1
					prev = line[b]
				} else {
					run++
				}
			}
			if run >= 5 {
				score += 3 + run - 5
			}
			for i := 0; i+10 < n; i++ {
				p := line[i : i+11]
				if pattern1011101(p[0:7]) && allFalse(p[7:11]) {
					score += 40
				}
				if pattern1011101(p[4:11]) && allFalse(p[0:4]) {
					score += 40
				}
			}
		}
	}
	for y := 0; y < n-1; y++ {
		for x := 0; x < n-1; x++ {
			v := m[y][x]
			if m[y][x+1] == v && m[y+1][x] == v && m[y+1][x+1] == v {
				score += 3
			}
		}
	}
	dark := 0
	for _, r := range m {
		for _, v := range r {
			if v {
				dark++
			}
		}
	}
	percent := dark * 100 / (n * n)
	score += absInt(percent-50) / 5 * 10
	return score
}
func pattern1011101(p []bool) bool {
	want := []bool{true, false, true, true, true, false, true}
	for i := range want {
		if p[i] != want[i] {
			return false
		}
	}
	return true
}
func allFalse(p []bool) bool {
	for _, b := range p {
		if b {
			return false
		}
	}
	return true
}

func maxIntQR(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// A centered logo must not overlap any function module, including alignment patterns.
func logoTouchesFunction(version, quiet int) bool {
	_, fn := qrBase(version)
	n := len(fn)
	side := float64(n+2*quiet) * .06
	left := (float64(n) - side) / 2
	right := (float64(n) + side) / 2
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			if float64(x) < right && float64(x+1) > left && float64(y) < right && float64(y+1) > left && fn[y][x] {
				return true
			}
		}
	}
	return false
}
