package etfapi

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// ErrBadPayload TODOC
var ErrBadPayload = errors.New("bad payload format")

// Payload TODOC
type Payload struct {
	OpCode    int
	Data      map[string]element
	SeqNum    *int
	EventName *string
}

// Marshal TODOC
func (p Payload) Marshal() []byte {
	return []byte{}
}

func (p *Payload) unmarshal(key string, val element) error {
	switch key {
	case "t":
		if val.code != 100 {
			return ErrBadPayload
		}
		eName := string(val.val)
		if eName != "nil" {
			p.EventName = &eName
		}

	case "s":
		if val.code != 98 && val.code != 100 {
			return ErrBadPayload
		}

		if val.code == 100 {
			eName := string(val.val)
			if eName != "nil" {
				return ErrBadPayload
			}
		} else {
			eVal := int(binary.BigEndian.Uint32(val.val))
			p.SeqNum = &eVal
		}

	case "op":
		if val.code != 97 {
			return ErrBadPayload
		}
		if len(val.val) != 1 {
			return ErrBadPayload
		}
		p.OpCode = int(val.val[0])
	case "d":
		if val.code != 116 {
			return ErrBadPayload
		}

		p.Data = map[string]element{}
		for i := 0; i < len(val.vals); i += 2 {
			if val.vals[i].code != 100 {
				return ErrBadPayload
			}

			p.Data[string(val.vals[i].val)] = val.vals[i+1]
		}
	default:
		return ErrBadPayload
	}
	return nil
}

type element struct {
	code byte
	val  []byte
	vals []element
}

func unmarshalSlice(raw []byte, numElements int) (uint32, []element, error) {
	var size uint32
	e := make([]element, numElements)
	var idx uint32
	var deltaIdx uint32
	var err error
	for i := 0; i < numElements; i++ {
		switch raw[idx] {
		case 116: // map
			e[i].code = 116
			idx++
			size = binary.BigEndian.Uint32(raw[idx : idx+4])
			idx += 4

			deltaIdx, e[i].vals, err = unmarshalSlice(raw[idx:], int(size*2))
			if err != nil {
				return 0, nil, err
			}
			idx += deltaIdx
		case 100: // atom
			e[i].code = 100
			idx++
			size = uint32(binary.BigEndian.Uint16(raw[idx : idx+2]))
			idx += 2
			e[i].val = raw[idx : idx+size]
			idx += size
		case 108: // list
			e[i].code = 108
			idx++
			size = binary.BigEndian.Uint32(raw[idx : idx+4])
			idx += 4
			deltaIdx, e[i].vals, err = unmarshalSlice(raw[idx:], int(size))
			if err != nil {
				return 0, nil, err
			}
			idx += deltaIdx

			if raw[idx] != 106 {
				return 0, nil, ErrBadPayload
			}
			idx++
		case 109: // binary
			e[i].code = 109
			idx++
			fmt.Println(raw[idx : idx+4])
			size = binary.BigEndian.Uint32(raw[idx : idx+4])
			idx += 4
			e[i].val = raw[idx : idx+size]
			idx += size
		case 98: // int
			e[i].code = 98
			idx++
			e[i].val = raw[idx : idx+4]
			idx += 4
		case 97: // small int
			e[i].code = 97
			idx++
			e[i].val = raw[idx : idx+1]
			idx++
		default:
			return 0, nil, fmt.Errorf("unknown field type %d", raw[idx])
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

	p := Payload{}

	_, eSlice, err := unmarshalSlice(raw[1:], 1)
	if err != nil {
		return nil, err
	}

	if eSlice[0].code != 116 { // not a map
		return nil, ErrBadPayload
	}

	if len(eSlice[0].vals)%2 != 0 {
		return nil, ErrBadPayload
	}

	for i := 0; i < len(eSlice[0].vals); i += 2 {
		p.unmarshal(string(eSlice[0].vals[i].val), eSlice[0].vals[i+1])
	}

	fmt.Printf("%+v\n", eSlice)

	fmt.Printf("unmarshal %d %+v\n", v, p)

	return &p, nil
}
