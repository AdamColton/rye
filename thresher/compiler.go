package thresher

import (
	"errors"
	"reflect"
	"strconv"
)

func (t *Thresher) compile(rt reflect.Type) uintPtrOp {
	switch rt.Kind() {
	case reflect.Ptr:
		rt = rt.Elem()
		return ptrMarshaller{
			op: t.compile(rt),
			t:  rt,
		}
	case reflect.Struct:
		return t.compileStruct(rt)
	case reflect.String:
		return uintPtrOpString{}
	case reflect.Int:
		return uintPtrOpInt{}
	case reflect.Int8:
		return uintPtrOpInt8{}
	case reflect.Int16:
		return uintPtrOpInt16C{}
	case reflect.Int32:
		return uintPtrOpInt32C{}
	case reflect.Int64:
		return uintPtrOpInt64C{}
	case reflect.Uint:
		return uintPtrOpUint{}
	case reflect.Uint8:
		return uintPtrOpByte{}
	case reflect.Uint16:
		return uintPtrOpUint16C{}
	case reflect.Uint32:
		return uintPtrOpUint32C{}
	case reflect.Uint64:
		return uintPtrOpUint64C{}
	case reflect.Float32:
		return uintPtrOpFloat32{}
	case reflect.Float64:
		return uintPtrOpFloat64{}
	case reflect.Slice:
		return t.compileSlice(rt.Elem())
	case reflect.Interface:
		return interfaceMarshaller{
			t:  t,
			rt: rt,
		}
	}
	return nil
}

func (t *Thresher) compileStruct(rt reflect.Type) *structMarshaller {
	if t.structMarshallers == nil {
		t.structMarshallers = make(map[reflect.Type]*structMarshaller)
	}
	if sm, found := t.structMarshallers[rt]; found {
		return sm
	}
	ln := rt.NumField()
	sm := &structMarshaller{
		byOrder: make([]structField, 0, ln),
	}
	t.structMarshallers[rt] = sm
	var max uint64
	for i := 0; i < ln; i++ {
		f := rt.Field(i)
		var skip bool
		idStr, found := f.Tag.Lookup("RyeField")
		if !found {
			skip = true
		}
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id == 0 {
			skip = true
		}
		if id > max {
			max = id
		}
		sf := structField{
			offset: f.Offset,
		}
		if skip {
			sf.uintPtrOp = uintPtrOpSkip{}
			sf.fieldHeader = 0
		} else {
			sf.uintPtrOp = t.compile(f.Type)
			sf.fieldHeader = id
		}
		sm.byOrder = append(sm.byOrder, sf)
	}
	sm.byId = make([]structField, max+1)
	for _, f := range sm.byOrder {
		if f.fieldHeader == 0 {
			continue
		}
		if sm.byId[f.fieldHeader].fieldHeader != 0 {
			panic(errors.New("RyeField redefined"))
		}
		sm.byId[f.fieldHeader] = f
	}
	return sm
}

func (t *Thresher) compileSlice(rt reflect.Type) sliceMarshaller {
	return sliceMarshaller{
		recordLen: rt.Size(),
		op:        t.compile(rt),
	}
}
