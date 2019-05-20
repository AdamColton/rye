package thresher

import (
	"fmt"
	//	"github.com/adamcolton/rye"
	"bytes"
	"encoding/gob"
	"github.com/stretchr/testify/assert"
	"strconv"
	"strings"
	"testing"
	"unsafe"
)

type Person struct {
	First        string `RyeField:"3"`
	Last         string `RyeField:"1"`
	Age          int    `RyeField:"2"`
	Role         int    `RyeField:"10"`
	StreetNumber int    `RyeField:"7"`
	StreetName   string `RyeField:"5"`
	City         string `RyeField:"6"`
}

func (p *Person) TypeID() uint64 {
	return 2
}

func TestRegister(t *testing.T) {
	p := &Person{
		First:        "Adam",
		Last:         "Colton",
		Age:          34,
		Role:         2,
		StreetNumber: 31415,
		StreetName:   "Pi Dr",
		City:         "Williamston",
	}

	p = nil

	th := &Thresher{}
	th.Register((*Person)(nil))

	b, err := th.Marshal(p, nil)
	assert.NoError(t, err)

	i, _, err := th.Unmarshal(b)
	assert.NoError(t, err)

	p2 := i.(*Person)
	assert.Equal(t, p, p2)
}

type Foo []string

func (*Foo) TypeID() uint64   { return 3 }
func (f *Foo) String() string { return strings.Join(*f, "|") }

func TestStringSlice(t *testing.T) {
	th := &Thresher{}
	th.Register((*Foo)(nil))

	f := &Foo{"this", "is", "a", "test"}
	b, err := th.Marshal(f, nil)
	assert.NoError(t, err)
	fmt.Println(b)

	i, _, err := th.Unmarshal(b)
	f2 := (i.(*Foo))
	assert.Equal(t, f, f2)
}

type AllTypes struct {
	Int       int          `RyeField:"1"`
	Int8      int8         `RyeField:"2"`
	Int16     int16        `RyeField:"3"`
	Int32     int32        `RyeField:"4"`
	Int64     int64        `RyeField:"5"`
	Uint      uint         `RyeField:"6"`
	Byte      byte         `RyeField:"7"`
	Uint8     uint8        `RyeField:"8"`
	Uint16    uint16       `RyeField:"9"`
	Uint32    uint32       `RyeField:"10"`
	Uint64    uint64       `RyeField:"11"`
	Float32   float32      `RyeField:"12"`
	Float64   float64      `RyeField:"13"`
	PtrInt    *int         `RyeField:"14"`
	Interface fmt.Stringer `RyeField:"15"`
}

func (*AllTypes) TypeID() uint64 {
	return 4
}

func TestAllTypes(t *testing.T) {
	th := &Thresher{}
	assert.NoError(t, th.Register((*AllTypes)(nil), (*Foo)(nil)))

	iPtr := 123
	ai := &AllTypes{
		Int:       1,
		Int8:      2,
		Int16:     3,
		Int32:     4,
		Int64:     5,
		Uint:      6,
		Byte:      7,
		Uint8:     8,
		Uint16:    9,
		Uint32:    10,
		Uint64:    11,
		Float32:   3.1415,
		Float64:   3.141592653,
		PtrInt:    &iPtr,
		Interface: &Foo{"a", "b", "c", "d"},
	}
	b, err := th.Marshal(ai, nil)
	assert.NoError(t, err)

	i, _, err := th.Unmarshal(b)
	assert.NoError(t, err)

	ai2 := i.(*AllTypes)

	assert.Equal(t, ai, ai2)
}

func TestAllTypesZero(t *testing.T) {
	th := &Thresher{}
	th.Register((*AllTypes)(nil))

	ai3 := &AllTypes{}
	b, err := th.Marshal(ai3, nil)
	assert.NoError(t, err)
	// 0: TypeID
	// 1: Ptr not null
	// 2: End
	assert.Len(t, b, 3)
}

type Bar struct {
	Foo string `RyeField:"1"`
	Bar int    `RyeField:"2"`
}

type BarSlice []Bar

func (*BarSlice) TypeID() uint64 { return 5 }

func TestSliceOfStruct(t *testing.T) {
	th := &Thresher{}
	th.Register((*BarSlice)(nil))

	bs := &BarSlice{
		Bar{"A", 1},
		Bar{"B", 2},
		Bar{"C", 3},
	}
	b, err := th.Marshal(bs, nil)
	assert.NoError(t, err)

	i, _, err := th.Unmarshal(b)
	bs2 := i.(*BarSlice)

	assert.Equal(t, bs, bs2)
}

type A struct {
	A int     `RyeField:"1"`
	B *B      `RyeField:"2"`
	C HasType `RyeField:"3"`
}

func (*A) TypeID() uint64 { return 6 }

type B struct {
	B int     `RyeField:"1"`
	A *A      `RyeField:"2"`
	C HasType `RyeField:"3"`
}

func (*B) TypeID() uint64 { return 7 }

func TestCyclic(t *testing.T) {
	th := &Thresher{}
	th.Register((*A)(nil))
	th.Register((*B)(nil))

	a := &A{
		A: 5,
		B: &B{
			B: 10,
		},
		C: &B{
			B: 15,
		},
	}
	b, err := th.Marshal(a, nil)
	assert.NoError(t, err)

	i, _, err := th.Unmarshal(b)
	a2 := i.(*A)

	assert.Equal(t, a, a2)
	a.B.B = 20
	assert.NotEqual(t, a, a2)
}

const (
	sflag uint64 = (1 << 63) - 1
)

func TestFloat(t *testing.T) {
	t.Skip() // looking into better way to compress a float
	f := 1.5
	u := *(*uint64)(unsafe.Pointer(&f))
	fmt.Println(b(u))
	s := u >> 63
	u &= sflag
	e := u >> 52
	m := u ^ (e << 52)
	fmt.Println(s, b(e), b(m))
}

func b(u uint64) string {
	s := strconv.FormatUint(u, 2)
	for len(s) < 64 {
		s = "0" + s
	}
	return s
}

func BenchmarkRye(b *testing.B) {
	th := &Thresher{}
	th.Register((*AllTypes)(nil), (*Foo)(nil))

	iPtr := 123
	ai := &AllTypes{
		Int:       1,
		Int8:      2,
		Int16:     3,
		Int32:     4,
		Int64:     5,
		Uint:      6,
		Byte:      7,
		Uint8:     8,
		Uint16:    9,
		Uint32:    10,
		Uint64:    11,
		Float32:   3.1415,
		Float64:   3.141592653,
		PtrInt:    &iPtr,
		Interface: &Foo{"a", "b", "c", "d"},
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		b, _ := th.Marshal(ai, nil)

		th.Unmarshal(b)
	}
}

func BenchmarkGob(b *testing.B) {
	iPtr := 123
	ai := &AllTypes{
		Int:       1,
		Int8:      2,
		Int16:     3,
		Int32:     4,
		Int64:     5,
		Uint:      6,
		Byte:      7,
		Uint8:     8,
		Uint16:    9,
		Uint32:    10,
		Uint64:    11,
		Float32:   3.1415,
		Float64:   3.141592653,
		PtrInt:    &iPtr,
		Interface: &Foo{"a", "b", "c", "d"},
	}
	buf := bytes.NewBuffer(nil)

	var out *AllTypes
	for n := 0; n < b.N; n++ {
		gob.NewEncoder(buf).Encode(ai)
		gob.NewDecoder(buf).Decode(out)
		buf.Reset()
	}
}
