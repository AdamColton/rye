package rye

const (
	endFlag      byte = 1 << 7
	sevenBitMask      = endFlag - 1
	endCheck          = uint64(endFlag)
	// CompactSize is the value used to indicate that a CompactUint64 should be
	// used when a size is specified.
	CompactSize int = 9
)

// CompactUint64 writes x to the Serializer in the Compact Uint64 format.
func (s *Serializer) CompactUint64(x uint64) {
	end := false
	for i := 0; !end && i < 8; i++ {
		end = x < endCheck
		b := byte(x) & sevenBitMask
		if end {
			b |= endFlag
		}
		s.Byte(b)
		x >>= 7
	}
	if !end {
		s.Byte(byte(x))
	}
}

func intToUint(x int64) uint64 {
	var sign uint64
	if x < 0 {
		sign = 1
		x = -(x + 1)
	}
	return uint64(x)<<1 | sign
}

func (s *Serializer) CompactInt64(x int64) {
	s.CompactUint64(intToUint(x))
}

// CompactUint64 reads a uint64 from the Deserializer in Compact Uint64 format.
func (d *Deserializer) CompactUint64() uint64 {
	var x uint64
	done := false
	for i := uint64(0); i < 8 && !done; i++ {
		b := d.Byte()
		done = b > sevenBitMask
		x += uint64(b&sevenBitMask) << (i * 7)
	}
	if !done {
		x += uint64(d.Byte()) << (8 * 7)
	}
	return x
}

func (d *Deserializer) CompactInt64() int64 {
	x := d.CompactUint64()
	i := int64(x >> 1)
	if x&1 == 1 {
		return -(i + 1)
	}
	return i
}

// CompactUint64Size returns the number of bytes needed to encode a uint64. It
// can take up to 10 bytes to encode a uint64.
func CompactUint64Size(x uint64) int {
	if x < (1 << (7 * 1)) {
		return 1 //1
	}

	if x < (1 << (7 * 3)) {
		if x < (1 << (7 * 2)) {
			return 2 //2
		}
		return 3 //2
	}

	if x < (1 << (7 * 6)) {
		if x < (1 << (7 * 4)) {
			return 4 //4
		}
		if x < (1 << (7 * 5)) {
			return 5 //5
		}
		return 6 //5
	}

	if x < (1 << (7 * 7)) {
		if x < (1 << (7 * 8)) {
			return 8 //5
		}
		return 7 //5
	}

	return 9 //4
}

func CompactInt64Size(x int64) int {
	return CompactUint64Size(intToUint(x))
}

func CompactSliceSize(b []byte) int {
	ln := len(b)
	return ln + int(CompactUint64Size(uint64(ln)))
}

func CompactStringSize(str string) int {
	return CompactSliceSize([]byte(str))
}
