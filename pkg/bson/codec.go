// Package bson provides BSON encoding/decoding for Atoms.
package bson

import (
	"github.com/zoobzio/atom"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Encode converts an Atom to a bson.M document.
func Encode(a *atom.Atom) (bson.M, error) {
	codec, ok := atom.CodecFor(a.Spec)
	if !ok {
		return nil, atom.ErrNoCodec
	}
	return bson.M(codec.EncodeMap(a)), nil
}

// Decode converts a bson.M document to an Atom.
func Decode(data bson.M, spec atom.Spec) (*atom.Atom, error) {
	codec, ok := atom.CodecFor(spec)
	if !ok {
		return nil, atom.ErrNoCodec
	}
	return codec.DecodeMap(convertBsonM(data))
}

// convertBsonM recursively converts bson.M to map[string]any.
func convertBsonM(m bson.M) map[string]any {
	result := make(map[string]any, len(m))
	for k, v := range m {
		switch val := v.(type) {
		case bson.M:
			result[k] = convertBsonM(val)
		case []any:
			result[k] = convertBsonSlice(val)
		default:
			result[k] = v
		}
	}
	return result
}

// convertBsonSlice recursively converts slice elements.
func convertBsonSlice(s []any) []any {
	result := make([]any, len(s))
	for i, v := range s {
		switch val := v.(type) {
		case bson.M:
			result[i] = convertBsonM(val)
		default:
			result[i] = v
		}
	}
	return result
}

// Marshal converts an Atom to BSON bytes.
func Marshal(a *atom.Atom) ([]byte, error) {
	m, err := Encode(a)
	if err != nil {
		return nil, err
	}
	return bson.Marshal(m)
}

// Unmarshal converts BSON bytes to an Atom.
func Unmarshal(data []byte, spec atom.Spec) (*atom.Atom, error) {
	var m bson.M
	if err := bson.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return Decode(m, spec)
}
