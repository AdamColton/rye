package compiler

import (
	"bytes"
	"encoding/gob"
	"github.com/stretchr/testify/assert"
	"testing"
)

type Person struct {
	First        string
	Last         string
	Age          int
	Role         int
	StreetNumber int
	StreetName   string
	City         string
	Data         []byte
}

func (*Person) Type() uint64 {
	return 1
}

var testPerson = Person{
	First:        "Lewis",
	Last:         "Carrol",
	Age:          34,
	Role:         123,
	StreetNumber: 31415,
	StreetName:   "Pi Drive",
	City:         "London",
	Data:         []byte{5, 4, 3, 2, 1},
}

func TestSize(t *testing.T) {
	s := Size(&testPerson)
	c := Compile((*Person)(nil))
	assert.Equal(t, s, c.Size(&testPerson))
}

func TestReflectMarshal(t *testing.T) {
	var out Person
	Unmarshal(Marshal(nil, &testPerson), &out)
	assert.Equal(t, testPerson, out)
}

func TestCompileMarshal(t *testing.T) {
	c := Compile((*Person)(nil))
	data := c.Marshal(nil, &testPerson)
	assert.Equal(t, Marshal(nil, &testPerson), data)
	var out Person
	c.Unmarshal(data, &out)
	assert.Equal(t, testPerson, out)
}

func BenchmarkReflectSize(b *testing.B) {
	for iter := 0; iter < b.N; iter++ {
		Size(&testPerson)
	}
}

func BenchmarkCompiledSize(b *testing.B) {
	c := Compile((*Person)(nil))
	for iter := 0; iter < b.N; iter++ {
		c.Size(&testPerson)
	}
}

func BenchmarkReflect(b *testing.B) {
	var person Person
	buf := make([]byte, Size(&testPerson))
	for iter := 0; iter < b.N; iter++ {
		Unmarshal(Marshal(buf, &testPerson), &person)
	}
}

func BenchmarkGob(b *testing.B) {
	var person Person
	buf := &bytes.Buffer{}
	for iter := 0; iter < b.N; iter++ {
		gob.NewEncoder(buf).Encode(&testPerson)
		gob.NewDecoder(buf).Decode(&person)
	}
}

func BenchmarkCompiled(b *testing.B) {
	c := Compile((*Person)(nil))
	buf := make([]byte, c.Size(&testPerson))
	var person Person
	for iter := 0; iter < b.N; iter++ {
		c.Unmarshal(c.Marshal(buf, &testPerson), &person)
	}
}
