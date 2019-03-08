// Package serial provides a number of helper for efficiently serializing data.
//
// Serialization is performed by the Serializer. It performs serialization in
// two passes. The first pass computes the size needed. The second pass writes
// the serialized data. After the size is set the Make method should be called
// to allocate the underlying byte slice. The entire slice is allocated once and
// and index to the current write position is tracked to reduce the amount of
// copying in the case of nested serialzing structures.
//
// Deserialization also tracks an index position so that the deserialization
// process mirrors the serialization process.
//
// One tool provided for serializing integer values is the Compact Uint64. This
// encodes seven bits per byte and uses the eighth bit as a continue flag. The
// trade off is that for some values, more bytes are used. But in cases where
// most of the vaules will be small, but occasionally large values will be
// encoded, this can be more efficient.
//
// Prefixers provide a useful tool to simplify the serialization process. If the
// data to be serialized can be represented as a slize of byte slices, the
// prefixer will handle the work of prefixing the length of each slice. The
// prefixer also guarentees the serializing and deserializing mirror each other.
package rye
