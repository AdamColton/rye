package rye

import (
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestLittleEdian(t *testing.T) {
	x16 := uint16(1234)
	s := Serializer{
		Data: make([]byte, 2),
	}
	s.Uint16(x16)
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, x16)
	assert.Equal(t, buf, s.Data)

	x32 := uint32(123456789)
	s = Serializer{
		Data: make([]byte, 4),
	}
	s.Uint32(x32)
	buf = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, x32)
	assert.Equal(t, buf, s.Data)

	x64 := uint64(123456789012345678)
	s = Serializer{
		Data: make([]byte, 8),
	}
	s.Uint64(x64)
	buf = make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, x64)
	assert.Equal(t, buf, s.Data)
}

func TestRoundTrip(t *testing.T) {
	s := Serializer{
		Data: make([]byte, 26),
	}

	d := Deserializer{
		Data: s.Data,
	}

	x16 := uint16(1234)
	x32 := uint32(123456789)
	x64 := uint64(123456789012345678)

	s.Uint16(x16)
	assert.Equal(t, x16, d.Uint16())
	s.Idx, d.Idx = 0, 0

	s.Uint32(x32)
	assert.Equal(t, x32, d.Uint32())
	s.Idx, d.Idx = 0, 0

	s.Uint64(x64)
	assert.Equal(t, x64, d.Uint64())
	s.Idx, d.Idx = 0, 0

	y16 := int16(-9234)
	y32 := int32(-923456789)
	y64 := int64(-923456789012345678)

	s.Int16(y16)
	assert.Equal(t, y16, d.Int16())
	s.Idx, d.Idx = 0, 0

	s.Int32(y32)
	assert.Equal(t, y32, d.Int32())
	s.Idx, d.Idx = 0, 0

	s.Int64(y64)
	assert.Equal(t, y64, d.Int64())
	s.Idx, d.Idx = 0, 0

	y16 = int16(234)
	y32 = int32(23456789)
	y64 = int64(23456789012345678)

	s.Int16(y16)
	assert.Equal(t, y16, d.Int16())
	s.Idx, d.Idx = 0, 0

	s.Int32(y32)
	assert.Equal(t, y32, d.Int32())
	s.Idx, d.Idx = 0, 0

	s.Int64(y64)
	assert.Equal(t, y64, d.Int64())
	s.Idx, d.Idx = 0, 0

	str := "Hi there, this is a test"
	s.CompactString(str)
	assert.Equal(t, str, d.CompactString())
	s.Idx, d.Idx = 0, 0

	var f32 float32 = 3.1415
	s.Float32(f32)
	assert.Equal(t, f32, d.Float32())
	s.Idx, d.Idx = 0, 0

	var f64 float64 = 3.1415926
	s.Float64(f64)
	assert.Equal(t, f64, d.Float64())
	s.Idx, d.Idx = 0, 0
}

func TestPrefixers(t *testing.T) {
	tt := []struct {
		name string
		data [][]byte
		p    Prefixer
	}{
		{
			name: "Static:Basic",
			data: [][]byte{
				{1, 2, 3, 4},
				{5, 6},
				{7, 8, 9, 10, 11, 12, 13, 14, 15},
				{16, 17, 18, 19, 20},
			},
			p: NewStaticPrefixer(1, 1, 1, 0),
		},
		{
			name: "Static:Compact",
			data: [][]byte{
				{1, 2, 3, 4},
				{5, 6},
				{7, 8, 9, 10, 11, 12, 13, 14, 15},
				{16, 17, 18, 19, 20},
			},
			p: NewStaticPrefixer(9, 9, 9, 0),
		},
		{
			name: "Dynamic:Basic",
			data: [][]byte{
				{1, 2, 3, 4},
				{5, 6},
				{7, 8, 9, 10, 11, 12, 13, 14, 15},
				{16, 17, 18, 19, 20},
			},
			p: NewDynamicPrefixer(1, 1),
		},
		{
			name: "Dynamic:Compact",
			data: [][]byte{
				{1, 2, 3, 4},
				{5, 6},
				{7, 8, 9, 10, 11, 12, 13, 14, 15},
				{16, 17, 18, 19, 20},
			},
			p: NewDynamicPrefixer(9, 9),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			s := &Serializer{}

			var err error
			s.Size, err = tc.p.Size(tc.data)
			assert.NoError(t, err)
			s.Make()

			s.Prefixer(tc.p, tc.data)
			d := &Deserializer{
				Data: s.Data,
			}
			assert.Equal(t, tc.data, d.Prefixer(tc.p))
		})
	}
}

type mockMarshaler struct {
	data []byte
}

func (m *mockMarshaler) MarshalSize() int {
	return len(m.data)
}

func (m *mockMarshaler) Marshal(s *Serializer) error {
	s.Slice(m.data)
	return nil
}

func (m *mockMarshaler) Unmarshal(d *Deserializer) error {
	m.data = d.Slice(len(d.Data))
	return nil
}

func TestMarshalUnmarshalHeader(t *testing.T) {
	m := &mockMarshaler{[]byte("this is a test")}
	headers := []HeaderSize{HeaderSize1, HeaderSize2, HeaderSize4, HeaderSize8, HeaderSizeCompact}

	for i, h := range headers {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			size := h.Size()
			if size == 10 {
				size = CompactUint64Size(uint64(len(m.data)))
			}
			s := &Serializer{
				Size: len(m.data) + size,
			}
			s.Make()
			assert.NoError(t, s.MarshalHeader(h, m))

			d := &Deserializer{
				Data: s.Data,
			}
			m2 := &mockMarshaler{}
			assert.NoError(t, d.UnmarshalHeader(h, m2))

			assert.Equal(t, m.data, m2.data)
		})
	}
}

func TestCompactUint64(t *testing.T) {
	tt := []uint64{
		0,
		1,
		(1 << 63),
		((1 << (63)) - 1) << 1,
	}
	for i := uint64(7); i < 63; i += 7 {
		tt = append(tt, (1<<i)-1, 1<<i, 1<<(i+1), (1<<(i+1))-1)
	}
	for _, tc := range tt {
		t.Run(strconv.FormatUint(tc, 10), func(t *testing.T) {
			assert.Equal(t, tc, e2eUint64(tc))
		})
	}

	// Explicitly test largest uint64
	var x uint64
	x--
	assert.Equal(t, x, e2eUint64(x))
}

func e2eUint64(x uint64) uint64 {
	s := &Serializer{
		Size: CompactUint64Size(x),
	}
	s.Make()

	s.CompactUint64(x)

	d := &Deserializer{
		Data: s.Data,
	}
	return d.CompactUint64()
}

func TestCompactInt64(t *testing.T) {
	tt := []int64{
		0,
		1,
		-1,
	}
	for i := int64(7); i < 62; i *= 2 {
		tt = append(tt, i, -i)
	}

	for _, tc := range tt {
		t.Run(strconv.FormatInt(tc, 10), func(t *testing.T) {
			assert.Equal(t, tc, e2eInt64(tc))
		})
	}
}

func e2eInt64(x int64) int64 {
	s := &Serializer{
		Size: CompactInt64Size(x),
	}
	s.Make()

	s.CompactInt64(x)

	d := &Deserializer{
		Data: s.Data,
	}
	return d.CompactInt64()
}

func TestSizeNegOne(t *testing.T) {
	assert.Equal(t, 1, CompactInt64Size(-1))
}

func TestCompactIntLargestSmallest(t *testing.T) {
	var l uint64
	l--
	l >>= 1
	i := int64(l)

	// i is the largest positive int that fits in int64; adding 1 rolls over to a
	// negative number
	assert.True(t, i+1 < i)

	s := &Serializer{
		Size: 9,
	}
	s.Make()
	s.CompactInt64(i)

	d := NewDeserializer(s.Data)
	i2 := d.CompactInt64()

	assert.Equal(t, i, i2)

	l = 1
	l <<= 63
	i = int64(l)

	// i is the largest negative int that fits in int64; subtracting one rolls
	// over to a positive number
	assert.True(t, i-1 > i)

	s = &Serializer{
		Size: 9,
	}
	s.Make()
	s.CompactInt64(i)

	d = NewDeserializer(s.Data)
	i2 = d.CompactInt64()

	assert.Equal(t, i, i2)
}
