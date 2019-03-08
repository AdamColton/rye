package rye

// Deserializer provides a helper for deserializing binary data
type Deserializer struct {
	Data []byte
	Idx  int
}

// NewDeserializer returns a Deserializer prepared to deserialize the provided
// data
func NewDeserializer(data []byte) *Deserializer {
	return &Deserializer{
		Data: data,
	}
}

// Sub returns a Sub-Deserializer of a given length. The index of the parent
// is placed at the end of the data allocated to the Sub-Deserializer.
func (d *Deserializer) Sub(ln int) *Deserializer {
	d.Idx += ln
	return &Deserializer{
		Data: d.Data[d.Idx-ln : d.Idx],
	}
}

// Byte returns one byte from the Deserializer and increases the index.
func (d *Deserializer) Byte() byte {
	d.Idx++
	return d.Data[d.Idx-1]
}

// Uint16 returns a uint16 from the Deserializer and increases the index.
func (d *Deserializer) Uint16() uint16 {
	d.Idx += 2
	return uint16(d.Data[d.Idx-1])<<8 + uint16(d.Data[d.Idx-2])
}

// Uint32 returns a uint32 from the Deserializer and increases the index.
func (d *Deserializer) Uint32() uint32 {
	d.Idx += 4
	return uint32(d.Data[d.Idx-1])<<24 +
		uint32(d.Data[d.Idx-2])<<16 +
		uint32(d.Data[d.Idx-3])<<8 +
		uint32(d.Data[d.Idx-4])
}

// Uint64 returns a uint64 from the Deserializer and increases the index.
func (d *Deserializer) Uint64() uint64 {
	d.Idx += 8
	return uint64(d.Data[d.Idx-1])<<56 +
		uint64(d.Data[d.Idx-2])<<48 +
		uint64(d.Data[d.Idx-3])<<40 +
		uint64(d.Data[d.Idx-4])<<32 +
		uint64(d.Data[d.Idx-5])<<24 +
		uint64(d.Data[d.Idx-6])<<16 +
		uint64(d.Data[d.Idx-7])<<8 +
		uint64(d.Data[d.Idx-8])
}

// Uint returns a uint64 from a specific number of bytes. The size must be
// either 1, 2, 4, 8 or 10 otherwise a 0 is returned. A value of 10 will make a
// call to CompactUint64
func (d *Deserializer) Uint(size int) uint64 {
	switch size {
	case 1:
		return uint64(d.Byte())
	case 2:
		return uint64(d.Uint16())
	case 4:
		return uint64(d.Uint32())
	case 8:
		return d.Uint64()
	case 10:
		return d.CompactUint64()
	}
	return 0
}

// Int16 returns an int16 from the Deserializer and increases the index.
func (d *Deserializer) Int16() int16 {
	return int16(d.Uint16())
}

// Int32 returns an int32 from the Deserializer and increases the index.
func (d *Deserializer) Int32() int32 {
	return int32(d.Uint32())
}

// Int64 returns an int64 from the Deserializer and increases the index.
func (d *Deserializer) Int64() int64 {
	return int64(d.Uint64())
}

// Float32 returns an float32 from the Deserializer and increases the index.
func (d *Deserializer) Float32() float32 {
	return uint32ToFloat32(d.Uint32())
}

// Float64 returns an float64 from the Deserializer and increases the index.
func (d *Deserializer) Float64() float64 {
	return uint64ToFloat64(d.Uint64())
}

// Slice returns a byte slice of the specified length and increases the index.
func (d *Deserializer) Slice(ln int) []byte {
	d.Idx += ln
	return d.Data[d.Idx-ln : d.Idx]
}

// String returns a string with a byte length of ln and increases the index.
func (d *Deserializer) String(ln int) string {
	return string(d.Slice(ln))
}

// UnmarshalHeader takes a HeaderSize to read the size of a Sub-Deserializer and
// passes the Sub-Deserializer into the unmarshaler. This is useful is the
// unmarshaler's final field is of an unspecified length.
func (d *Deserializer) UnmarshalHeader(headerBytes HeaderSize, unmarshaler Unmarshaler) error {
	ln := int(d.Uint(headerBytes.Size()))
	sub := d.Sub(ln)
	return unmarshaler.Unmarshal(sub)
}
