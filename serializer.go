package rye

// Serializer is used to Serialize into the Data field.
type Serializer struct {
	Data []byte
	Size int
	Idx  int
}

// Make will set Data to the length of Size. If Data is already populated, it
// will be appeded to.
func (s *Serializer) Make() {
	ln := s.Size - len(s.Data)
	if ln > 0 {
		s.Data = append(s.Data, make([]byte, ln)...)
	}
}

// Byte writes a byte to the Serializer and increases the index.
func (s *Serializer) Byte(b byte) {
	s.Data[s.Idx] = b
	s.Idx++
}

// Uint16 writes a uint16 to the Serializer and increases the index.
func (s *Serializer) Uint16(x uint16) {
	s.Data[s.Idx] = byte(x)
	s.Data[s.Idx+1] = byte(x >> 8)
	s.Idx += 2
}

// Uint32 writes a uint32 to the Serializer and increases the index.
func (s *Serializer) Uint32(x uint32) {
	s.Data[s.Idx] = byte(x)
	s.Data[s.Idx+1] = byte(x >> 8)
	s.Data[s.Idx+2] = byte(x >> 16)
	s.Data[s.Idx+3] = byte(x >> 24)
	s.Idx += 4
}

// Uint64 writes a uint64 to the Serializer and increases the index.
func (s *Serializer) Uint64(x uint64) {
	s.Data[s.Idx] = byte(x)
	s.Data[s.Idx+1] = byte(x >> 8)
	s.Data[s.Idx+2] = byte(x >> 16)
	s.Data[s.Idx+3] = byte(x >> 24)
	s.Data[s.Idx+4] = byte(x >> 32)
	s.Data[s.Idx+5] = byte(x >> 40)
	s.Data[s.Idx+6] = byte(x >> 48)
	s.Data[s.Idx+7] = byte(x >> 56)
	s.Idx += 8
}

// Uint writes the value value to the Serializer using the specified size. Size
// must be 1, 2, 4, 8 or 10. A value of 10 will use a Compact Uint64, all other
// vaules will use the specified number of bytes.
func (s *Serializer) Uint(size int, value uint64) {
	switch size {
	case 1:
		s.Byte(byte(value))
	case 2:
		s.Uint16(uint16(value))
	case 4:
		s.Uint32(uint32(value))
	case 8:
		s.Uint64(uint64(value))
	case 10:
		s.CompactUint64(value)
	}
}

// Int16 writes a int16 to the Serializer and increases the index.
func (s *Serializer) Int16(x int16) {
	s.Uint16(uint16(x))
}

// Int32 writes a int32 to the Serializer and increases the index.
func (s *Serializer) Int32(x int32) {
	s.Uint32(uint32(x))
}

// Int64 writes a int64 to the Serializer and increases the index.
func (s *Serializer) Int64(x int64) {
	s.Uint64(uint64(x))
}

// Float32 writes a float32 to the Serializer and increases the index.
func (s *Serializer) vFloat32(f float32) {
	s.Uint32(float32ToUint32(f))
}

// Float64 writes a float64 to the Serializer and increases the index.
func (s *Serializer) Float64(f float64) {
	s.Uint64(float64ToUint64(f))
}

// Slice writes a byte slice to the Serializer and increases the index.
func (s *Serializer) Slice(data []byte) {
	copy(s.Data[s.Idx:], data)
	s.Idx += len(data)
}

// String writes a string to the Serializer and increases the index.
func (s *Serializer) String(data string) {
	s.Slice([]byte(data))
}

// MarshalHeader will take a marshaller and prepend it's size. Useful when
// serializing a collection.
func (s *Serializer) MarshalHeader(headerBytes HeaderSize, marshaler Marshaler) error {
	s.Uint(headerBytes.Size(), uint64(marshaler.MarshalSize()))
	return marshaler.Marshal(s)
}
