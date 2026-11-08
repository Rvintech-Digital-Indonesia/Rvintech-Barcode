package qr

// QR encoding uses the ISO/IEC 18004 Numeric, Alphanumeric, and Byte modes.
// Dynamic programming considers every legal segment boundary and minimizes
// encoded bits, including each mode and character-count header.
type segment struct {
	mode int
	data []byte
}

const (
	numericMode = 1
	alphaMode   = 2
	byteMode    = 4
)
const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ $%*+-./:"

func countWidth(mode, version int) int {
	group := 0
	if version >= 10 && version <= 26 {
		group = 1
	}
	if version >= 27 {
		group = 2
	}
	switch mode {
	case numericMode:
		return [...]int{10, 12, 14}[group]
	case alphaMode:
		return [...]int{9, 11, 13}[group]
	default:
		return [...]int{8, 16, 16}[group]
	}
}
func payloadBits(mode, n int) int {
	switch mode {
	case numericMode:
		return n/3*10 + [...]int{0, 4, 7}[n%3]
	case alphaMode:
		return n/2*11 + n%2*6
	default:
		return n * 8
	}
}
func optimizeSegments(data []byte, version int) ([]segment, int) {
	n := len(data)
	inf := int(^uint(0)>>1) / 4
	cost := make([]int, n+1)
	prev := make([]int, n+1)
	pmode := make([]int, n+1)
	for i := 1; i <= n; i++ {
		cost[i] = inf
	}
	for i := 0; i < n; i++ {
		if cost[i] == inf {
			continue
		}
		for _, mode := range []int{numericMode, alphaMode, byteMode} {
			cw := countWidth(mode, version)
			maxCount := (1 << cw) - 1
			for j := i + 1; j <= n; j++ {
				l := j - i
				if l > maxCount {
					break
				}
				valid := true
				for _, b := range data[i:j] {
					if mode == numericMode && (b < '0' || b > '9') {
						valid = false
						break
					}
					if mode == alphaMode && alphaIndex(b) < 0 {
						valid = false
						break
					}
				}
				if !valid {
					break
				}
				c := cost[i] + 4 + cw + payloadBits(mode, l)
				if c < cost[j] {
					cost[j] = c
					prev[j] = i
					pmode[j] = mode
				}
			}
		}
	}
	if n == 0 {
		return nil, 0
	}
	segs := []segment{}
	for end := n; end > 0; {
		start := prev[end]
		segs = append(segs, segment{pmode[end], append([]byte(nil), data[start:end]...)})
		end = start
	}
	for i, j := 0, len(segs)-1; i < j; i, j = i+1, j-1 {
		segs[i], segs[j] = segs[j], segs[i]
	}
	return segs, cost[n]
}
func alphaIndex(b byte) int {
	for i := 0; i < len(alphabet); i++ {
		if alphabet[i] == b {
			return i
		}
	}
	return -1
}
func encodeSegments(segs []segment, capacity, version int) []byte {
	bits := make([]bool, 0, capacity*8)
	put := func(v uint, n int) {
		for i := n - 1; i >= 0; i-- {
			bits = append(bits, (v>>i)&1 != 0)
		}
	}
	for _, s := range segs {
		cw := countWidth(s.mode, version)
		put(uint(s.mode), 4)
		put(uint(len(s.data)), cw)
		switch s.mode {
		case numericMode:
			for i := 0; i < len(s.data); {
				n := len(s.data) - i
				if n > 3 {
					n = 3
				}
				v := 0
				for _, b := range s.data[i : i+n] {
					v = v*10 + int(b-'0')
				}
				put(uint(v), [...]int{0, 4, 7, 10}[n])
				i += n
			}
		case alphaMode:
			for i := 0; i < len(s.data); i += 2 {
				a := alphaIndex(s.data[i])
				if i+1 < len(s.data) {
					put(uint(a*45+alphaIndex(s.data[i+1])), 11)
				} else {
					put(uint(a), 6)
				}
			}
		default:
			for _, b := range s.data {
				put(uint(b), 8)
			}
		}
	}
	for i := 0; i < 4 && len(bits) < capacity*8; i++ {
		bits = append(bits, false)
	}
	for len(bits)%8 != 0 {
		bits = append(bits, false)
	}
	out := make([]byte, 0, capacity)
	for i := 0; i < len(bits); i += 8 {
		var b byte
		for j := 0; j < 8; j++ {
			b <<= 1
			if bits[i+j] {
				b |= 1
			}
		}
		out = append(out, b)
	}
	pad := byte(0xEC)
	for len(out) < capacity {
		out = append(out, pad)
		pad ^= 0xFD
	}
	return out
}
