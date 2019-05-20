package thresher

import (
	"errors"
	"github.com/adamcolton/rye"
	"reflect"
	"unsafe"
)

type Thresher struct {
	typedIDMarshallers []*marshaller
	structMarshallers  map[reflect.Type]*structMarshaller
}

func (t *Thresher) Unmarshal(data []byte) (interface{}, map[uint64]interface{}, error) {
	d := rye.NewDeserializer(data)
	vt := int(d.CompactUint64())
	if vt > len(t.typedIDMarshallers) {
		return nil, nil, errors.New("Not found")
	}
	m := t.typedIDMarshallers[vt]
	if m == nil {
		return nil, nil, errors.New("Not found")
	}

	r := reflect.New(m.t)
	i := r.Elem().Interface()
	base := uintptr(unsafe.Pointer(&i)) + ifcePtrOffset
	m.op.unmarshal(uintptr(unsafe.Pointer(base)), d)
	return i, nil, nil
}

func (t *Thresher) Marshal(v HasType, in []byte) ([]byte, error) {
	vt := v.TypeID()
	if len(t.typedIDMarshallers) < int(vt) {
		return nil, errors.New("Not found")
	}
	m := t.typedIDMarshallers[vt]
	if m == nil {
		return nil, errors.New("not found")
	}

	base := uintptr(unsafe.Pointer(&v)) + ifcePtrOffset

	s := &rye.Serializer{}
	if in == nil {
		s.Data = make([]byte, m.op.size(base)+rye.CompactUint64Size(vt))
	} else {
		s.Size = len(in)
		s.Data = in
	}
	s.CompactUint64(vt)
	m.op.marshal(base, s)
	return s.Data, nil
}

func (t *Thresher) Register(vs ...HasType) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	for _, v := range vs {
		vid := v.TypeID()
		if len(t.typedIDMarshallers) <= int(vid) {
			ln := int(vid)
			if ln < 256 {
				ln = 256
			}
			s := make([]*marshaller, ln)
			if t.typedIDMarshallers != nil {
				copy(s, t.typedIDMarshallers)
			}
			t.typedIDMarshallers = s
		}
		if t.typedIDMarshallers[vid] != nil {
			return errors.New("TypeID redefined")
		}
		vt := reflect.TypeOf(v)
		t.typedIDMarshallers[vid] = &marshaller{
			op: t.compile(vt),
			t:  vt,
		}
	}
	return
}
