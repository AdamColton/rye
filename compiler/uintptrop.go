package compiler

import (
	"github.com/adamcolton/rye"
	"unsafe"
)

type uintPtrOp interface {
	size(base uintptr, offset uintptr) int
	marshal(base uintptr, offset uintptr, s *rye.Serializer)
	unmarshal(base uintptr, offset uintptr, d *rye.Deserializer)
}

type uintPtrOpByteSlice struct{}

func (uintPtrOpByteSlice) size(base uintptr, offset uintptr) int {
	ln := len(*(*[]byte)(unsafe.Pointer(base + offset)))
	return ln + rye.CompactUint64Size(uint64(ln))
}

func (uintPtrOpByteSlice) marshal(base uintptr, offset uintptr, s *rye.Serializer) {
	b := *(*[]byte)(unsafe.Pointer(base + offset))
	s.CompactUint64(uint64(len(b)))
	s.Slice(b)
}
func (uintPtrOpByteSlice) unmarshal(base uintptr, offset uintptr, d *rye.Deserializer) {
	ln := int(d.CompactUint64())
	b := (*[]byte)(unsafe.Pointer(base + offset))
	*b = d.Slice(ln)
}

type uintPtrOpString struct{}

func (uintPtrOpString) size(base uintptr, offset uintptr) int {
	ln := len(*(*string)(unsafe.Pointer(base + offset)))
	return ln + rye.CompactUint64Size(uint64(ln))
}

func (uintPtrOpString) marshal(base uintptr, offset uintptr, s *rye.Serializer) {
	str := *(*string)(unsafe.Pointer(base + offset))
	s.CompactUint64(uint64(len(str)))
	s.String(str)
}
func (uintPtrOpString) unmarshal(base uintptr, offset uintptr, d *rye.Deserializer) {
	ln := int(d.CompactUint64())
	str := (*string)(unsafe.Pointer(base + offset))
	*str = d.String(ln)
}

type uintPtrOpInt struct{}

func (uintPtrOpInt) size(base uintptr, offset uintptr) int {
	return rye.CompactUint64Size(*(*uint64)(unsafe.Pointer(base + offset)))
}

func (uintPtrOpInt) marshal(base uintptr, offset uintptr, s *rye.Serializer) {
	u := *(*int)(unsafe.Pointer(base + offset))
	s.CompactUint64(uint64(u))
}
func (uintPtrOpInt) unmarshal(base uintptr, offset uintptr, d *rye.Deserializer) {
	u := int(d.CompactUint64())
	ptr := (*int)(unsafe.Pointer(base + offset))
	*ptr = u
}
