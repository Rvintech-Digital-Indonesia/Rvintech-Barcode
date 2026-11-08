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

// minTree stores the best prefix cost at each quotient index. Separate trees
// per residue make the numeric and alphanumeric payload cost separable into a
// prefix-dependent term and a constant for the current endpoint.
type minTree struct {
	size  int
	cost  []int
	index []int
}

func newMinTree(n int) *minTree {
	size := 1
	for size < n {
		size <<= 1
	}
	t := &minTree{size: size, cost: make([]int, size*2), index: make([]int, size*2)}
	for i := range t.cost {
		t.cost[i] = int(^uint(0)>>1) / 4
		t.index[i] = -1
	}
	return t
}

func (t *minTree) update(pos, cost, index int) {
	p := pos + t.size
	if cost >= t.cost[p] {
		return
	}
	t.cost[p], t.index[p] = cost, index
	for p >>= 1; p > 0; p >>= 1 {
		left, right := p*2, p*2+1
		if t.cost[left] <= t.cost[right] {
			t.cost[p], t.index[p] = t.cost[left], t.index[left]
		} else {
			t.cost[p], t.index[p] = t.cost[right], t.index[right]
		}
	}
}

func (t *minTree) query(left, right int) (int, int) {
	best, index := int(^uint(0)>>1)/4, -1
	for left, right = left+t.size, right+t.size; left < right; left, right = left>>1, right>>1 {
		if left&1 != 0 {
			if t.cost[left] < best {
				best, index = t.cost[left], t.index[left]
			}
			left++
		}
		if right&1 != 0 {
			right--
			if t.cost[right] < best {
				best, index = t.cost[right], t.index[right]
			}
		}
	}
	return best, index
}

func optimizeSegments(data []byte, version int) ([]segment, int) {
	n := len(data)
	if n == 0 {
		return nil, 0
	}
	cwNumeric, cwAlpha, cwByte := countWidth(numericMode, version), countWidth(alphaMode, version), countWidth(byteMode, version)
	maxNumeric, maxAlpha, maxByte := (1<<cwNumeric)-1, (1<<cwAlpha)-1, (1<<cwByte)-1
	// Consecutive valid run lengths let each endpoint restrict queries without
	// rescanning substrings.
	numericRun, alphaRun := make([]int, n+1), make([]int, n+1)
	for i := 0; i < n; i++ {
		if data[i] >= '0' && data[i] <= '9' {
			numericRun[i+1] = numericRun[i] + 1
		}
		if alphaIndex(data[i]) >= 0 {
			alphaRun[i+1] = alphaRun[i] + 1
		}
	}
	inf := int(^uint(0)>>1) / 4
	cost := make([]int, n+1)
	prev := make([]int, n+1)
	pmode := make([]int, n+1)
	for i := 1; i <= n; i++ {
		cost[i] = inf
	}
	// For each mode and residue, maintain min(cost[i] - slope*floor(i/modulus)).
	// Range minimum queries yield the exact optimal predecessor in O(log n).
	trees := [3][3]*minTree{}
	moduli := [3]int{3, 2, 1}
	for m := range trees {
		for r := 0; r < moduli[m]; r++ {
			trees[m][r] = newMinTree(n/moduli[m] + 2)
		}
	}
	insert := func(i int) {
		if cost[i] == inf {
			return
		}
		trees[0][i%3].update(i/3, cost[i]-10*(i/3), i)
		trees[1][i%2].update(i/2, cost[i]-11*(i/2), i)
		trees[2][0].update(i, cost[i]-8*i, i)
	}
	insert(0)
	for j := 1; j <= n; j++ {
		consider := func(mode, treeMode, m, cw, maxCount, run int) {
			low := j - maxCount
			if minRun := j - run; minRun > low {
				low = minRun
			}
			if low < 0 {
				low = 0
			}
			if low >= j {
				return
			}
			for residue := 0; residue < m; residue++ {
				if j-1 < residue {
					continue
				}
				qLow := low / m
				if low%m > residue {
					qLow++
				}
				qHigh := (j - 1 - residue) / m
				if qLow > qHigh {
					continue
				}
				_, i := trees[treeMode][residue].query(qLow, qHigh+1)
				if i < 0 {
					continue
				}
				c := cost[i] + 4 + cw + payloadBits(mode, j-i)
				if c < cost[j] {
					cost[j], prev[j], pmode[j] = c, i, mode
				}
			}
		}
		consider(numericMode, 0, 3, cwNumeric, maxNumeric, numericRun[j])
		consider(alphaMode, 1, 2, cwAlpha, maxAlpha, alphaRun[j])
		consider(byteMode, 2, 1, cwByte, maxByte, j)
		insert(j)
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
