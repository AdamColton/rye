package compiler

import (
	"fmt"
	"github.com/adamcolton/rye"
	"reflect"
	"unsafe"
)

type StructMarshaller interface {
	Size(HasType) int
	Marshal([]byte, HasType) []byte
	Unmarshal([]byte, HasType)
	private()
}

type structField struct {
	offset uintptr
	uintPtrOp
}

type structFields []structField

func (s structFields) Size(i HasType) int {
	base := *(*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(&i)) + 8))
	var size int
	for _, f := range s {
		size += f.size(base, f.offset)
	}
	return size
}

func (sf structFields) Marshal(in []byte, i HasType) []byte {
	base := *(*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(&i)) + 8))
	s := &rye.Serializer{}
	if in == nil {
		for _, f := range sf {
			s.Size += f.size(base, f.offset)
		}
		s.Data = make([]byte, s.Size)
	} else {
		s.Size = len(in)
		s.Data = in
	}
	for _, f := range sf {
		f.marshal(base, f.offset, s)
	}
	return s.Data
}

func (sf structFields) Unmarshal(data []byte, i HasType) {
	base := *(*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(&i)) + 8))
	d := rye.NewDeserializer(data)
	for _, f := range sf {
		f.unmarshal(base, f.offset, d)
	}
}

func (structFields) private() {}

func compileStruct(t reflect.Type) structFields {
	ln := t.NumField()
	s := make(structFields, 0, ln)
	for i := 0; i < ln; i++ {
		f := t.Field(i)
		sf := structField{
			offset: f.Offset,
		}
		switch f.Type.Kind() {
		case reflect.String:
			sf.uintPtrOp = uintPtrOpString{}
			s = append(s, sf)
		case reflect.Int:
			sf.uintPtrOp = uintPtrOpInt{}
			s = append(s, sf)
		case reflect.Slice:
			switch f.Type.Elem().Kind() {
			case reflect.Uint8:
				sf.uintPtrOp = uintPtrOpByteSlice{}
				s = append(s, sf)
			}
		}
	}
	return s
}

type HasType interface {
	Type() uint64
}

func Compile(i HasType) StructMarshaller {
	t := reflect.TypeOf(i)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Struct:
		return compileStruct(t)
	}
	fmt.Println("Cannot compile", t)
	return nil
}
