package rye

// Marshaler requires that an object to be marshaled knows how many bytes it
// will use and know how to marhsal itself into a Serializer.
type Marshaler interface {
	MarshalSize() int
	Marshal(*Serializer) error
}

// Unmarshaler provides an interfaces for unmarshalling using a Deserializer.
type Unmarshaler interface {
	Unmarshal(*Deserializer) error
}

// Marshal creates a Serializer, sets it size from MarshalSize, allocates the
// slice and calls Marhsal.
func Marshal(marshaler Marshaler) ([]byte, error) {
	s := &Serializer{
		Size: marshaler.MarshalSize(),
	}
	s.Make()
	return s.Data, marshaler.Marshal(s)
}
