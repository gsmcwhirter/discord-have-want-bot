package etfapi

import (
	"bytes"
	"fmt"
	"io"

	"github.com/pkg/errors"

	"github.com/gsmcwhirter/eso-discord/pkg/discordapi/constants"
)

// ErrNotImplemented TODOC
var ErrNotImplemented = errors.New("not yet implemented")

// ErrBadTarget TODOC
var ErrBadTarget = errors.New("bad element unmarshal target")

// ErrBadPayload TODOC
var ErrBadPayload = errors.New("bad payload format")

// ErrBadFieldType TODOC
var ErrBadFieldType = errors.New("bad field type")

// ErrBadMarshalData TODOC
var ErrBadMarshalData = errors.New("bad marshal data")

// ErrBadElementData TODOC
var ErrBadElementData = errors.New("bad element data")

// ETFCode TODOC
type ETFCode int

// TODOC
const (
	Map    ETFCode = 116
	Atom           = 100
	List           = 108
	Binary         = 109
	Int32          = 98
	Int8           = 97
)

// Element TODOC
type Element struct {
	Code ETFCode
	Val  []byte
	Vals []Element
}

// NewElement TODOC
func NewElement(code ETFCode, val interface{}) (*Element, error) {
	e := &Element{
		Code: code,
	}

	if v, ok := val.([]Element); ok {
		if code != Map && code != List {
			return nil, ErrBadElementData
		}
		e.Vals = v

		return e, nil
	}

	if v, ok := val.([]byte); ok {
		if code == Map || code == List {
			return nil, ErrBadElementData
		}
		e.Val = v

		return e, nil
	}

	return e, ErrBadElementData
}

// WriteTo TODOC
func (e *Element) WriteTo(b io.Writer) (int64, error) {
	var tmp interface{}
	if e.Val != nil {
		tmp = e.Val
	} else if e.Vals != nil {
		tmp = e.Vals
	} else {
		tmp = nil
	}

	data, err := marshalInterface(e.Code, tmp)
	if err != nil {
		return 0, errors.Wrap(err, "couldn't marshal element")
	}

	n, err := b.Write(data)
	return int64(n), err
}

// Unmarshal TODOC
func (e *Element) Unmarshal(target interface{}) error {
	var err error
	switch e.Code {
	case Atom:
		v, ok := target.(*string)
		if !ok {
			return errors.Wrap(ErrBadTarget, "needed *string")
		}

		*v = string(e.Val)

		return nil

	case Binary:
		v, ok := target.([]byte)
		if !ok {
			return errors.Wrap(ErrBadTarget, "needed []byte")
		}

		if len(v) < len(e.Val) {
			return errors.Wrap(ErrBadTarget, "target buffer too small")
		}

		copy(v, e.Val)

		return nil
	case Int32:
		v, ok := target.(*int)
		if !ok {
			return errors.Wrap(ErrBadTarget, "needed *int")
		}

		*v, err = Int32SliceToInt(e.Val)

		return errors.Wrap(err, "could not unmarshal int32")
	case Int8:
		v, ok := target.(*int)
		if !ok {
			return errors.Wrap(ErrBadTarget, "needed *int")
		}

		*v, err = Int8SliceToInt(e.Val)

		return errors.Wrap(err, "could not unmarshal int8")
	case Map:
		v, ok := target.(map[string]Element)
		if !ok {
			return errors.Wrap(ErrBadTarget, "needed map[string]Element")
		}

		if len(e.Vals)%2 != 0 {
			return ErrBadParity
		}

		for i := 0; i < len(e.Vals); i += 2 {
			if e.Vals[i].Code != Atom {
				return ErrBadFieldType
			}

			v[string(e.Vals[i].Val)] = e.Vals[i+1]
		}

		return nil
	case List:
		return ErrNotImplemented
	default:
		return ErrBadElementData
	}
}

// Payload TODOC
type Payload struct {
	OpCode    constants.OpCode
	Data      map[string]Element
	SeqNum    *int
	EventName *string
}

func (p Payload) String() string {
	return fmt.Sprintf("Payload{OpCode=%d, Data=%+v, SeqNum=%v, EventName=%v}", p.OpCode, p.Data, p.SeqNum, p.EventName)
}

// Marshal code

// Marshal TODOC
func (p Payload) Marshal() ([]byte, error) {
	var data []byte
	var err error

	b := bytes.Buffer{}
	b.WriteByte(131)

	len := 2
	if p.SeqNum != nil {
		len++
	}

	err = b.WriteByte(byte(Map))
	if err != nil {
		return nil, errors.Wrap(err, "unable to write outer map label")
	}
	err = writeLength32(&b, len)
	if err != nil {
		return nil, errors.Wrap(err, "unable to write outer map length")
	}

	err = writeAtom(&b, []byte("d"))
	if err != nil {
		return nil, errors.Wrap(err, "unable to write 'd' key")
	}
	data, err = marshalInterface(Map, ElementMapToElementSlice(p.Data))
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal 'd' value")
	}
	_, err = b.Write(data)
	if err != nil {
		return nil, errors.Wrap(err, "unable to write 'd' value")
	}

	err = writeAtom(&b, []byte("op"))
	if err != nil {
		return nil, errors.Wrap(err, "unable to write 'op' key")
	}
	data, err = IntToInt8Slice(int(p.OpCode))
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert 'op' value to byte slice")
	}
	data, err = marshalInterface(Int8, data)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal 'op' value")
	}
	_, err = b.Write(data)
	if err != nil {
		return nil, errors.Wrap(err, "unable to write 'op' value")
	}

	if p.SeqNum != nil {
		err = writeAtom(&b, []byte("s"))
		if err != nil {
			return nil, errors.Wrap(err, "unable to write 's' key")
		}
		data, err = IntToInt32Slice(*p.SeqNum)
		if err != nil {
			return nil, errors.Wrap(err, "unable to convert 's' value to byte slice")
		}
		data, err = marshalInterface(Int32, data)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal 's' value")
		}
		_, err = b.Write(data)
		if err != nil {
			return nil, errors.Wrap(err, "unable to write 's' value")
		}
	}

	return b.Bytes(), nil
}

func writeAtom(b *bytes.Buffer, val []byte) error {
	err := b.WriteByte(byte(Atom))
	if err != nil {
		return errors.Wrap(err, "could not write label")
	}

	size, err := IntToInt16Slice(len(val))
	if err != nil {
		return errors.Wrap(err, "couldn't marshal size")
	}

	_, err = b.Write(size)
	if err != nil {
		return errors.Wrap(err, "could not write size")
	}

	_, err = b.Write(val)
	if err != nil {
		return errors.Wrap(err, "could not write value")
	}

	return nil
}

func writeLength32(b io.Writer, n int) error {
	size, err := IntToInt32Slice(n)
	if err != nil {
		return errors.Wrap(err, "could not marshal length")
	}

	_, err = b.Write(size)
	return errors.Wrap(err, "could not write length")
}

func marshalInterface(code ETFCode, val interface{}) ([]byte, error) {
	// var data []byte
	var err error

	b := bytes.Buffer{}
	b.WriteByte(byte(code))

	switch code {
	case Map:
		v, ok := val.([]Element)
		if !ok {
			return nil, errors.Wrap(ErrBadMarshalData, "not a list of elements")
		}

		if len(v)%2 != 0 {
			return nil, errors.Wrap(ErrBadMarshalData, "bad parity on map list")
		}

		err = writeLength32(&b, len(v)/2)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't marshal map length")
		}

		for i := 0; i < len(v); i += 2 {
			if v[i].Code != Atom {
				return nil, errors.Wrap(err, "bad map key")
			}

			err = writeAtom(&b, v[i].Val)
			if err != nil {
				return nil, errors.Wrap(err, "couldn't marshal map key")
			}

			_, err = v[i+1].WriteTo(&b)
			if err != nil {
				return nil, errors.Wrap(err, "couldn't marshal map value")
			}
		}
	case List:
		v, ok := val.([]Element)
		if !ok {
			return nil, errors.Wrap(ErrBadMarshalData, "not a list of elements")
		}

		err = writeLength32(&b, len(v))
		if err != nil {
			return nil, errors.Wrap(err, "couldn't marshal list length")
		}

		for _, e := range v {
			_, err = e.WriteTo(&b)
			if err != nil {
				return nil, errors.Wrap(err, "couldn't marshal list value")
			}
		}

		err = b.WriteByte(106)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't write trailing list byte")
		}

	case Binary:
		v, ok := val.([]byte)
		if !ok {
			return nil, errors.Wrap(ErrBadMarshalData, "not a byte slice")
		}

		err = writeLength32(&b, len(v))
		if err != nil {
			return nil, errors.Wrap(err, "couldn't marshal binary length")
		}

		_, err = b.Write(v)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't marshal binary value")
		}

	case Atom:
		v, ok := val.([]byte)
		if !ok {
			return nil, errors.Wrap(ErrBadMarshalData, "not a byte slice")
		}

		err = writeAtom(&b, v)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't marshal Atom value")
		}
	case Int32:
		v, ok := val.([]byte)
		if !ok {
			return nil, errors.Wrap(ErrBadMarshalData, "not a byte slice")
		}

		if len(v) != 4 {
			return nil, errors.Wrap(ErrBadMarshalData, "not a int32 byte slice")
		}

		_, err = b.Write(v)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't marshal Int32 value")
		}

	case Int8:
		v, ok := val.([]byte)
		if !ok {
			return nil, errors.Wrap(ErrBadMarshalData, "not a byte slice")
		}

		if len(v) != 1 {
			return nil, errors.Wrap(ErrBadMarshalData, "not a int8 byte slice")
		}

		_, err = b.Write(v)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't marshal Int8 value")
		}

	default:
		return nil, ErrBadMarshalData
	}

	return b.Bytes(), nil
}

// UnMarshal code

func (p *Payload) unmarshal(key string, val Element) error {
	var err error

	switch key {
	case "t":
		if val.Code != Atom {
			return ErrBadPayload
		}

		var eName string
		err = val.Unmarshal(&eName)
		if err != nil {
			return errors.Wrap(err, "bad payload")
		}

		if eName != "nil" {
			p.EventName = &eName
		}

	case "s":
		if val.Code != Int32 && val.Code != Atom { // Atom for "nil"
			return ErrBadPayload
		}

		if val.Code == Atom {
			var eName string
			err = val.Unmarshal(&eName)
			if err != nil || eName != "nil" {
				return ErrBadPayload
			}

		} else {
			var eVal int
			err = val.Unmarshal(&eVal)
			if err != nil {
				return errors.Wrap(err, "bad payload")
			}

			p.SeqNum = &eVal
		}

	case "op":
		if val.Code != Int8 {
			return ErrBadPayload
		}

		var eVal int
		err = val.Unmarshal(&eVal)
		if err != nil {
			return errors.Wrap(err, "bad payload")
		}
		p.OpCode = constants.OpCode(eVal)

	case "d":
		if val.Code != Map {
			return ErrBadPayload
		}

		p.Data = map[string]Element{}
		err = val.Unmarshal(p.Data)
		if err != nil {
			return errors.Wrap(err, "bad payload")
		}
	default:
		return ErrBadPayload
	}
	return nil
}

func unmarshalSlice(raw []byte, numElements int) (uint32, []Element, error) {
	var size int
	var idx uint32
	var deltaIdx uint32
	var err error

	e := make([]Element, numElements)

	for i := 0; i < numElements; i++ {
		e[i].Code = ETFCode(raw[idx])
		idx++
		switch e[i].Code {
		case Map:
			size, err = Int32SliceToInt(raw[idx : idx+4])
			if err != nil {
				return 0, nil, errors.Wrap(err, "could not read map length")
			}
			idx += 4

			deltaIdx, e[i].Vals, err = unmarshalSlice(raw[idx:], size*2)
			if err != nil {
				return 0, nil, errors.Wrap(err, "could not unmarshal map")
			}
			idx += deltaIdx
		case Atom:
			size, err = Int16SliceToInt(raw[idx : idx+2])
			if err != nil {
				return 0, nil, errors.Wrap(err, "could not read atom length")
			}
			idx += 2
			e[i].Val = raw[idx : idx+uint32(size)]
			idx += uint32(size)
		case List:
			size, err = Int32SliceToInt(raw[idx : idx+4])
			if err != nil {
				return 0, nil, errors.Wrap(err, "coult not read list length")
			}
			idx += 4
			deltaIdx, e[i].Vals, err = unmarshalSlice(raw[idx:], size)
			if err != nil {
				return 0, nil, err
			}
			idx += deltaIdx

			if raw[idx] != 106 {
				return 0, nil, ErrBadPayload
			}
			idx++
		case Binary:
			size, err = Int32SliceToInt(raw[idx : idx+4])
			if err != nil {
				return 0, nil, errors.Wrap(err, "could not read binary length")
			}
			idx += 4
			e[i].Val = raw[idx : idx+uint32(size)]
			idx += uint32(size)
		case Int32:
			e[i].Val = raw[idx : idx+4]
			idx += 4
		case Int8: // small int
			e[i].Val = raw[idx : idx+1]
			idx++
		default:
			return 0, nil, ErrBadFieldType
		}
	}

	return idx, e, nil
}

// Unmarshal TODOC
func Unmarshal(raw []byte) (*Payload, error) {
	if len(raw) < 2 {
		return nil, ErrBadPayload
	}
	v := int(raw[0])
	if v != 131 {
		return nil, ErrBadPayload
	}

	p := Payload{}

	_, eSlice, err := unmarshalSlice(raw[1:], 1)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal bytes")
	}

	if eSlice[0].Code != 116 { // not a map
		return nil, errors.Wrap(ErrBadPayload, "payload not a map")
	}

	if len(eSlice[0].Vals)%2 != 0 {
		return nil, errors.Wrap(ErrBadPayload, "map parity incorrect incomplete")
	}

	for i := 0; i < len(eSlice[0].Vals); i += 2 {
		err = p.unmarshal(string(eSlice[0].Vals[i].Val), eSlice[0].Vals[i+1])
		if err != nil {
			return nil, errors.Wrap(err, "could not unmarshal field")
		}
	}

	return &p, nil
}
