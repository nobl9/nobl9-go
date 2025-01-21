package serdeutil

import "github.com/pkg/errors"

// RawMessage is a raw encoded JSON or YAML value.
// It implements:
//   - [json.Marshaler] and [json.Unmarshaler]
//   - [yaml.BytesMarshaler] and [yaml.BytesUnmarshaler]
//
// It can be used to delay JSON/YAML decoding or precompute a JSON/YAML encoding.
type RawMessage []byte

func (m RawMessage) MarshalJSON() ([]byte, error) { return m.marshal() }
func (m RawMessage) MarshalYAML() ([]byte, error) { return m.marshal() }

func (m RawMessage) marshal() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	return m, nil
}

func (m *RawMessage) UnmarshalJSON(data []byte) error { return m.unmarshal(data) }
func (m *RawMessage) UnmarshalYAML(data []byte) error { return m.unmarshal(data) }

func (m *RawMessage) unmarshal(data []byte) error {
	if m == nil {
		return errors.New("RawMessage: unmarshal on nil pointer")
	}
	*m = append((*m)[0:0], data...)
	return nil
}
