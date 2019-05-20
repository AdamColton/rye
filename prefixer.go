package rye

import (
	"fmt"
)

// Prefixer provides a strategy for encoding a [][]byte to a []byte.
type Prefixer interface {
	Size([][]byte) (int, error)
	Deserialize(*Deserializer) [][]byte
	Serialize(*Serializer, [][]byte) error
	private()
}

type staticPrefixer struct {
	size    int
	headers []int
}

// NewStaticPrefixer returns a Prefixer that will prefix a set number of byte
// slices. One header is given for each slice that will be prefixed. A header
// value of 1,2,4,6 or 8 indicates the number of bytes to use for the header. A
// value of 9 indicates that a Compact uint64 should be used. A negative value
// indicates that no header should be included and the length will be the
// positive value. If the last header is 0, that indicates that no header should
// be included because the remainder of the data is part of the final slice. So
// NewStaticPrefixer(2, 9, -6, 0) says there will be four byte slices. For the
// first, two bytes should be used express the length, for the second a compact
// uint64 should be used, the third slice will be exactly 6 bytes long and the
// last will have no header and should read to the end of the serialized data.
// An invalid header value will cause NewStaticPrefixer to panic.
func NewStaticPrefixer(headers ...int) Prefixer {
	var size int
	for i, h := range headers {
		if h > 0 {
			if h != 1 && h != 2 && h != 4 && h != 8 && h != CompactSize {
				panic("Positive header size must be 1, 2, 4, 8 or 9")
			}
			if h != CompactSize {
				size += h
			}
		} else if h == 0 && i != len(headers)-1 {
			panic("Header value of 0 only valid in last position")
		}
	}

	return &staticPrefixer{
		size:    size,
		headers: headers,
	}
}

func (*staticPrefixer) private() {}

func (p *staticPrefixer) Size(data [][]byte) (int, error) {
	if len(p.headers) != len(data) {
		return 0, ErrSizeMismatch{len(p.headers), len(data)}
	}
	sum := p.size
	for i, d := range data {
		if p.headers[i] == CompactSize {
			sum += CompactUint64Size(uint64(len(d)))
		}
		sum += len(d)
	}
	return sum, nil
}

// ErrSizeMismatch will be returned when a Static Serializer is used and the
// number of slices does not match the number of headers.
type ErrSizeMismatch struct {
	Expected, Got int
}

func (e ErrSizeMismatch) Error() string {
	return fmt.Sprintf("Prefixer size mismatch; expected: %d got: %d", e.Expected, e.Got)
}

func (p *staticPrefixer) Serialize(s *Serializer, data [][]byte) error {
	if len(p.headers) != len(data) {
		return ErrSizeMismatch{len(p.headers), len(data)}
	}

	for i, hln := range p.headers {
		if hln > 0 {
			if hln == CompactSize {
				s.CompactUint64(uint64(len(data[i])))
			} else {
				s.Uint(hln, uint64(len(data[i])))
			}
		}
		s.Slice(data[i])
	}

	return nil
}

func (p *staticPrefixer) Deserialize(d *Deserializer) [][]byte {
	out := make([][]byte, len(p.headers))
	var ln int

	for i, hln := range p.headers {
		if hln == 0 {
			out[i] = d.Data[d.Idx:]
			d.Idx += len(out[i])
			return out
		} else if hln > 0 {
			if hln == CompactSize {
				ln = int(d.CompactUint64())
			} else {
				ln = int(d.Uint(hln))
			}
		} else {
			ln = -hln
		}

		out[i] = d.Slice(ln)
	}

	return out
}

type dynamicPrefixer struct {
	outer, inner int
}

// NewDynamicPrefixer writes one outer prefix that holds the number of slices,
// then each slice has it's own prefix holding it's length. For both outer and
// inner length bytes 1, 2, 4 or 8 indicate the number of bytes to use to encode
// the length. A value of 9 indicates that a Compact Uint64 should be used. A
// negative value indicates that the length is fixed to the positive value. So
// NewDynamicPrefixer(1,2) would use one byte to encode the outer length and 2
// bytes to encdoe the length of each byte slice. NewDynamicPrefixer(9, -32)
// would use a Compact Uint64 to encode the number of slices and each slice
// would be exactly 32 bytes long
func NewDynamicPrefixer(outerLengthBytes, innerLengthBytes int) Prefixer {
	if outerLengthBytes >= 0 && outerLengthBytes != 1 && outerLengthBytes != 2 && outerLengthBytes != 4 && outerLengthBytes != 8 && outerLengthBytes != CompactSize {
		panic("outerLengthBytes must be 1, 2, 4, 8, 9 or negative")
	}

	if innerLengthBytes >= 0 && innerLengthBytes != 1 && innerLengthBytes != 2 && innerLengthBytes != 4 && innerLengthBytes != 8 && innerLengthBytes != CompactSize {
		panic("innerLengthBytes must be 1, 2, 4, 8, 9 or negative")
	}

	return &dynamicPrefixer{
		outer: outerLengthBytes,
		inner: innerLengthBytes,
	}
}

func (p *dynamicPrefixer) Size(data [][]byte) (int, error) {
	var size int
	if p.outer > 0 {
		if p.outer == CompactSize {
			size = CompactUint64Size(uint64(len(data)))
		} else {
			size = p.outer
		}
	}
	var h int
	if p.inner > 0 {
		h = p.inner
	}

	for _, b := range data {
		if h == CompactSize {
			size += CompactUint64Size(uint64(len(b))) + len(b)
		} else {
			size += h + len(b)
		}
	}
	return size, nil
}

func (p *dynamicPrefixer) Deserialize(d *Deserializer) [][]byte {
	var outer int
	if p.outer > 0 {
		if p.outer == CompactSize {
			outer = int(d.CompactUint64())
		} else {
			outer = int(d.Uint(p.outer))
		}
	} else {
		outer = -p.outer
	}

	var inner int
	data := make([][]byte, outer)
	for i := range data {
		if p.inner > 0 {
			if p.inner == CompactSize {
				inner = int(d.CompactUint64())
			} else {
				inner = int(d.Uint(p.inner))
			}
		} else {
			inner = -p.inner
		}
		data[i] = d.Slice(inner)
	}
	return data
}

func (p *dynamicPrefixer) Serialize(s *Serializer, data [][]byte) error {
	if p.outer > 0 {
		if p.outer == CompactSize {
			s.CompactUint64(uint64(len(data)))
		} else {
			s.Uint(p.outer, uint64(len(data)))
		}
	}

	for _, b := range data {
		if p.inner > 0 {
			if p.outer == CompactSize {
				s.CompactUint64(uint64(len(b)))
			} else {
				s.Uint(p.inner, uint64(len(b)))
			}
		}
		s.Slice(b)
	}

	return nil
}

func (p *dynamicPrefixer) private() {}

// Prefixer takes in an instance of a Prefixer and uses it to deserialize a
// [][]byte
func (d *Deserializer) Prefixer(pre Prefixer) [][]byte {
	return pre.Deserialize(d)
}

// Prefixer serializes the data with provided Prefixer.
func (s *Serializer) Prefixer(pre Prefixer, data [][]byte) error {
	return pre.Serialize(s, data)
}
