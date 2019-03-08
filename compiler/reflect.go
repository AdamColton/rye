package compiler

import (
	"github.com/adamcolton/rye"
	"reflect"
)

func Size(i interface{}) int {
	var size int
	v := reflect.ValueOf(i)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return -1
	}

	nf := v.NumField()
	for i := 0; i < nf; i++ {
		f := v.Field(i)
		switch f.Kind() {
		case reflect.String:
			ln := f.Len()
			size += rye.CompactUint64Size(uint64(ln)) + ln
		case reflect.Int, reflect.Int16:
			size += rye.CompactUint64Size(uint64(f.Int()))
		case reflect.Slice:
			switch f.Type().Elem().Kind() {
			case reflect.Uint8:
				ln := f.Len()
				size += rye.CompactUint64Size(uint64(ln)) + ln
			}
		}
	}
	return size
}

func Marshal(in []byte, i interface{}) []byte {
	s := &rye.Serializer{}
	if in == nil {
		s.Size = Size(i)
		s.Data = make([]byte, s.Size)
	} else {
		s.Size = len(in)
		s.Data = in
	}

	v := reflect.ValueOf(i)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	nf := v.NumField()
	for i := 0; i < nf; i++ {
		f := v.Field(i)
		switch f.Kind() {
		case reflect.String:
			str := f.String()
			s.CompactUint64(uint64(len(str)))
			s.String(str)
		case reflect.Int, reflect.Int16:
			s.CompactUint64(uint64(f.Int()))
		case reflect.Slice:
			switch f.Type().Elem().Kind() {
			case reflect.Uint8:
				d := f.Interface().([]byte)
				s.CompactUint64(uint64(len(d)))
				s.Slice(d)
			}
		}
	}
	return s.Data
}

func Unmarshal(data []byte, i interface{}) {
	d := rye.NewDeserializer(data)

	v := reflect.ValueOf(i)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return
	}

	nf := v.NumField()
	for i := 0; i < nf; i++ {
		f := v.Field(i)
		switch f.Kind() {
		case reflect.String:
			ln := int(d.CompactUint64())
			f.SetString(d.String(ln))
		case reflect.Int, reflect.Int16:
			f.SetInt(int64(d.CompactUint64()))
		case reflect.Slice:
			switch f.Type().Elem().Kind() {
			case reflect.Uint8:
				ln := int(d.CompactUint64())
				f.Set(reflect.ValueOf(d.Slice(ln)))
			}
		}
	}
}
