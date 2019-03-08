package rye

// HeaderSize provides a way to specify only valid header lengths to
// MarshalHeader and UnmarshalHeader.
type HeaderSize interface {
	Size() int
	private()
}

type priv struct{}

func (priv) private() {}

type headerSize1 struct{ priv }

func (headerSize1) Size() int { return 1 }

// HeaderSize1 specifies a header size of 1 byte
var HeaderSize1 = headerSize1{}

type headerSize2 struct{ priv }

func (headerSize2) Size() int { return 2 }

// HeaderSize2 specifies a header size of 2 bytes
var HeaderSize2 = headerSize2{}

type headerSize4 struct{ priv }

func (headerSize4) Size() int { return 4 }

// HeaderSize4 specifies a header size of 4 bytes
var HeaderSize4 = headerSize4{}

type headerSize8 struct{ priv }

func (headerSize8) Size() int { return 8 }

// HeaderSize8 specifies a header size of 8 bytes
var HeaderSize8 = headerSize8{}

type headerSizeCompact struct{ priv }

func (headerSizeCompact) Size() int { return CompactSize }

// HeaderSizeCompact specifies a header that uses Compact Uint64
var HeaderSizeCompact = headerSizeCompact{}
