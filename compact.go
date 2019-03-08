package rye

const (
	endFlag      byte = 1 << 7
	sevenBitMask      = endFlag - 1
	endCheck          = uint64(endFlag)
	// CompactSize is the value used to indicate that a CompactUint64 should be
	// used when a size is specified.
	CompactSize int = 10
)

// CompactUint64 writes x to the Serializer in the Compact Uint64 format.
func (s *Serializer) CompactUint64(x uint64) {
	end := false
	for !end {
		end = x < endCheck
		b := byte(x) & sevenBitMask
		if end {
			b |= endFlag
		}
		s.Byte(b)
		x >>= 7
	}
}

// CompactUint64 reads a uint64 from the Deserializer in Compact Uint64 format.
func (d *Deserializer) CompactUint64() uint64 {
	var x uint64
	done := false
	for i := uint64(0); i < 10 && !done; i++ {
		b := d.Byte()
		done = b > sevenBitMask
		x += uint64(b&sevenBitMask) << (i * 7)
	}
	return x
}

// CompactUint64Size returns the number of bytes needed to encode a uint64. It
// can take up to 10 bytes to encode a uint64.
func CompactUint64Size(x uint64) int {
	if x < (1 << (7 * 1)) {
		return 1 //1
	}

	if x < (1 << (7 * 3)) {
		if x < (1 << (7 * 2)) {
			return 2 //3
		}
		return 3 //3
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

	if x < (1 << (7 * 9)) {
		if x < (1 << (7 * 7)) {
			return 7 //5
		}
		if x < (1 << (7 * 8)) {
			return 8 //6
		}
		return 9 //6
	}

	return 10 //4
}
