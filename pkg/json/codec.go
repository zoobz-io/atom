// Package json provides JSON encoding/decoding for Atoms.
package json //nolint:revive // intentional name conflict for clear API

import (
	"encoding/json"

	"github.com/zoobzio/atom"
)

// Encode converts an Atom to JSON bytes.
func Encode(a *atom.Atom) ([]byte, error) {
	codec, ok := atom.CodecFor(a.Spec)
	if !ok {
		return nil, atom.ErrNoCodec
	}
	return json.Marshal(codec.EncodeMap(a))
}

// EncodeIndent converts an Atom to indented JSON bytes.
func EncodeIndent(a *atom.Atom, prefix, indent string) ([]byte, error) {
	codec, ok := atom.CodecFor(a.Spec)
	if !ok {
		return nil, atom.ErrNoCodec
	}
	return json.MarshalIndent(codec.EncodeMap(a), prefix, indent)
}

// Decode converts JSON bytes to an Atom.
func Decode(data []byte, spec atom.Spec) (*atom.Atom, error) {
	codec, ok := atom.CodecFor(spec)
	if !ok {
		return nil, atom.ErrNoCodec
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return codec.DecodeMap(m)
}
