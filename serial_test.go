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
		Data: make([]byte, 8),
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
}

func TestPrefixers(t *testing.T) {
	tt := []struct {
		name string
		data [][]byte
		p    Prefixer
	}{
		{
			name: "Static - Basic",
			data: [][]byte{
				{1, 2, 3, 4},
				{5, 6},
				{7, 8, 9, 10, 11, 12, 13, 14, 15},
				{16, 17, 18, 19, 20},
			},
			p: NewStaticPrefixer(1, 1, 1, 0),
		},
		{
			name: "Static - Compact",
			data: [][]byte{
				{1, 2, 3, 4},
				{5, 6},
				{7, 8, 9, 10, 11, 12, 13, 14, 15},
				{16, 17, 18, 19, 20},
			},
			p: NewStaticPrefixer(10, 10, 10, 0),
		},
		{
			name: "Dynamic - Basic",
			data: [][]byte{
				{1, 2, 3, 4},
				{5, 6},
				{7, 8, 9, 10, 11, 12, 13, 14, 15},
				{16, 17, 18, 19, 20},
			},
			p: NewDynamicPrefixer(1, 1),
		},
		{
			name: "Dynamic - Compact",
			data: [][]byte{
				{1, 2, 3, 4},
				{5, 6},
				{7, 8, 9, 10, 11, 12, 13, 14, 15},
				{16, 17, 18, 19, 20},
			},
			p: NewDynamicPrefixer(10, 10),
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

func TestCompact(t *testing.T) {
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
			s := &Serializer{
				Size: CompactUint64Size(tc),
			}
			s.Make()

			s.CompactUint64(tc)

			d := &Deserializer{
				Data: s.Data,
			}

			assert.Equal(t, tc, d.CompactUint64())
		})
	}
}
