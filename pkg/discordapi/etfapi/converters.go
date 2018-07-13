package etfapi

import (
	"encoding/binary"

	"github.com/pkg/errors"
)

// ErrOutOfBounds TODOC
var ErrOutOfBounds = errors.New("int value out of bounds")

// ErrBadParity TODOC
var ErrBadParity = errors.New("non-even list parity")

// IntToInt8Slice TODOC
func IntToInt8Slice(v int) ([]byte, error) {
	if v < 0 || v > 255 {
		return nil, ErrOutOfBounds
	}

	return []byte{byte(v)}, nil
}

// Int8SliceToInt TODOC
func Int8SliceToInt(v []byte) (int, error) {
	if len(v) != 1 {
		return 0, ErrOutOfBounds
	}

	return int(v[0]), nil
}

// IntToInt16Slice TODOC
func IntToInt16Slice(v int) ([]byte, error) {
	if v < 0 || v >= (1<<16) {
		return nil, ErrOutOfBounds
	}

	size := make([]byte, 2)
	binary.BigEndian.PutUint16(size, uint16(v))

	return size, nil
}

// Int16SliceToInt TODOC
func Int16SliceToInt(v []byte) (int, error) {
	if len(v) != 2 {
		return 0, ErrOutOfBounds
	}

	return int(binary.BigEndian.Uint16(v)), nil
}

// IntToInt32Slice TODOC
func IntToInt32Slice(v int) ([]byte, error) {
	if v < 0 || v >= (1<<32) {
		return nil, ErrOutOfBounds
	}

	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, uint32(v))

	return size, nil
}

// Int32SliceToInt TODOC
func Int32SliceToInt(v []byte) (int, error) {
	if len(v) != 4 {
		return 0, ErrOutOfBounds
	}

	return int(binary.BigEndian.Uint32(v)), nil
}

// ElementMapToElementSlice TODOC
func ElementMapToElementSlice(m map[string]Element) []Element {
	e := make([]Element, 0, len(m)*2)
	for k, v := range m {
		e = append(e, Element{
			Code: Atom,
			Val:  []byte(k),
		})
		e = append(e, v)
	}

	return e
}
