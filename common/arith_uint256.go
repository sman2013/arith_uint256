package common

import (
	"encoding/binary"
)

type ArithUint256 struct {
	pn [8]uint32
}

// FromUint64
func (a *ArithUint256) FromUint64(v uint64) {
	a.pn[0] = uint32(v)
	a.pn[1] = uint32(v >> 32)
	for i := 2; i < 7; i++ {
		a.pn[i] = 0
	}
}

func reverse(s []byte) []byte {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// FromHash
func (a *ArithUint256) FromHash(h Hash) {
	for i := 0; i < 8; i++ {
		a.pn[i] = 0
	}
	hBytes := h.Bytes()
	for i := 0; i < 8; i++ {
		buf := hBytes[i*4 : (i+1)*4]
		buf = reverse(buf)
		a.pn[7-i] = binary.LittleEndian.Uint32(buf)
	}
}

// SetCompact set nBits
func (a *ArithUint256) SetCompact(v uint32) {
	nSize := v >> 24
	nWord := v & 0x007fffff
	if nSize <= 3 {
		nWord >>= 8 * (3 - nSize)
		a.FromUint64(uint64(nWord))
	} else {
		a.FromUint64(uint64(nWord))
		a.LeftShift(8 * (nSize - 3))
	}
}

// GetCompact get nBits
func (a *ArithUint256) GetCompact() uint32 {
	aBak := a.Copy()
	nSize := (a.bits() + 7) / 8
	var nCompact uint32
	if nSize <= 3 {
		nCompact = uint32(a.getLow64() << (8 * uint32(3-nSize)))
	} else {
		aBak.RightShift(8 * (nSize - 3))
		nCompact = uint32(aBak.getLow64())
	}
	if nCompact&0x00800000 != 0 {
		nCompact >>= 8
		nSize++
	}
	nCompact |= uint32(nSize) << 24
	return nCompact
}

// GetBytes
func (a *ArithUint256) GetBytes() []byte {
	res := make([]byte, 0)
	for i := 0; i < 8; i++ {
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, a.pn[i])
		buf = reverse(buf)
		res = append(buf, res...)
	}
	return res
}

func (a *ArithUint256) getLow64() uint64 {
	return uint64(a.pn[0]) | uint64(a.pn[1])<<32
}

func (a *ArithUint256) Copy() *ArithUint256 {
	b := new(ArithUint256)
	for i := 0; i < 8; i++ {
		b.pn[i] = a.pn[i]
	}
	return b
}

// LeftShift a << shift
func (a *ArithUint256) LeftShift(shift uint32) {
	pnBak := a.pn
	for i := 0; i < 8; i++ {
		a.pn[i] = 0
	}
	k := shift / 32
	shift %= 32
	for i := uint32(0); i < 8; i++ {
		if i+k+1 < 8 && shift != 0 {
			a.pn[i+k+1] |= pnBak[i] >> (32 - shift)
		}
		if i+k < 8 {
			a.pn[i+k] |= pnBak[i] << shift
		}
	}
}

// RightShift a >> shift
func (a *ArithUint256) RightShift(shift int) {
	pnBak := a.pn
	for i := 0; i < 8; i++ {
		a.pn[i] = 0
	}
	k := shift / 32
	shift %= 32
	for i := 0; i < 8; i++ {
		if i-k-1 >= 0 && shift != 0 {
			a.pn[i-k-1] |= pnBak[i] << uint32(32-shift)
		}
		if i-k >= 0 {
			a.pn[i-k] |= pnBak[i] >> uint32(shift)
		}
	}
}

// MulU32
func (a *ArithUint256) MulU32(b uint32) {
	var carry uint64
	for i := 0; i < 8; i++ {
		n := carry + uint64(b)*uint64(a.pn[i])
		a.pn[i] = uint32(n & 0xffffffff)
		carry = n >> 32
	}
}

// MulU64
func (a *ArithUint256) MulU64(v uint64) {
	b := new(ArithUint256)
	b.FromUint64(v)
	a.Mul(b)
}

// Mul a *= b
func (a *ArithUint256) Mul(b *ArithUint256) {
	tmp := new(ArithUint256)
	for j := 0; j < 8; j++ {
		carry := uint64(0)
		for i := 0; i+j < 8; i++ {
			n := carry + uint64(tmp.pn[i+j]) + uint64(a.pn[j])*uint64(b.pn[i])
			tmp.pn[i+j] = uint32(n & 0xffffffff)
			carry = n >> 32
		}
	}
	a.pn = tmp.pn
}

// DivU64
func (a *ArithUint256) DivU64(v uint64) {
	b := new(ArithUint256)
	b.FromUint64(v)
	a.Div(b)
}

// Div a /= b
func (a *ArithUint256) Div(b *ArithUint256) {
	div := b.Copy()
	num := a.Copy()
	for i := 0; i < 8; i++ {
		a.pn[i] = 0
	}
	numBits := num.bits()
	divBits := div.bits()
	if divBits == 0 {
		panic("Division by zero")
	}
	if divBits > numBits {
		return
	}
	shift := numBits - divBits
	div.LeftShift(uint32(shift))
	for shift >= 0 {
		if num.Cmp(div) >= 0 {
			num.Sub(div)
			a.pn[shift/32] |= 1 << uint32(shift&31)
		}
		div.RightShift(1) // shift back
		shift--
	}
}

func (a *ArithUint256) bits() int {
	for pos := 7; pos >= 0; pos-- {
		if a.pn[pos] > 0 {
			for nbits := 31; nbits > 0; nbits-- {
				if (a.pn[pos] & (1 << uint32(nbits))) > 0 {
					return 32*pos + nbits + 1
				}
			}
			return 32*pos + 1
		}
	}
	return 0
}

// Cmp a>b return 1, a==b return 0, a<b return -1
func (a *ArithUint256) Cmp(b *ArithUint256) int {
	for i := 7; i >= 0; i-- {
		if a.pn[i] < b.pn[i] {
			return -1
		}
		if a.pn[i] > b.pn[i] {
			return 1
		}
	}
	return 0
}

// Sub a -= b
func (a *ArithUint256) Sub(b *ArithUint256) {
	a.Add(b.Neg())
}

// Add a += b
func (a *ArithUint256) Add(b *ArithUint256) {
	var carry uint64
	for i := 0; i < 8; i++ {
		n := carry + uint64(a.pn[i]+b.pn[i])
		a.pn[i] = uint32(n & 0xffffffff)
		carry = n >> 32
	}
}

// Add1 a += 1
func (a *ArithUint256) Add1() {
	for i := 0; i < 8; i++ {
		a.pn[i]++
		if a.pn[i] == 0 {
			i++
		} else {
			break
		}
	}
}

// Not ~a or ^a
func (a *ArithUint256) Not() *ArithUint256 {
	b := new(ArithUint256)
	for i := 0; i < 8; i++ {
		b.pn[i] = ^(a.pn[i])
	}
	return b
}

// Neg -a
func (a *ArithUint256) Neg() *ArithUint256 {
	b := new(ArithUint256)
	for i := 0; i < 8; i++ {
		b.pn[i] = ^(a.pn[i])
	}
	b.Add1()
	return b
}

// And a &= b
func (a *ArithUint256) And(b *ArithUint256) {
	for i := 0; i < 8; i++ {
		a.pn[i] &= b.pn[i]
	}
}

// Or a |= b
func (a *ArithUint256) Or(b *ArithUint256) {
	for i := 0; i < 8; i++ {
		a.pn[i] |= b.pn[i]
	}
}

// Xor a ^= b
func (a *ArithUint256) Xor(b *ArithUint256) {
	for i := 0; i < 8; i++ {
		a.pn[i] ^= b.pn[i]
	}
}
