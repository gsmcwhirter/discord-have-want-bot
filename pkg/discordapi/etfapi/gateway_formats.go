package etfapi

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/gsmcwhirter/eso-discord/pkg/util"
)

// ErrBadPayload TODOC
var ErrBadPayload = errors.New("bad payload format")

// Payload TODOC
type Payload struct {
	OpCode    int
	Data      []byte
	SeqNum    int
	EventName string
}

// Marshal TODOC
func (p Payload) Marshal() []byte {
	return []byte{}
}

func (p *Payload) unmarshal(key string, val []byte) error {
	fmt.Printf("*** %s => %v\n", key, val)
	switch key {
	case "t":
		p.EventName = string(val)
	case "s":
		// p.SeqNum = int()
	case "op":
		// p.OpCode = int(fval)
	case "d":
		p.Data = append(make([]byte, len(val)), val...)
	case "heartbeat_interval":
	case "_trace":
	default:
		return ErrBadPayload
	}
	return nil
}

// Unmarshal TODOC
func Unmarshal(raw []byte) (*Payload, error) {
	if len(raw) < 2 {
		return nil, ErrBadPayload
	}
	v := int(raw[0])
	sep := raw[1]

	p := Payload{}

	// TODO: here
	var b *bytes.Buffer
	var r io.ReadCloser
	var fname []byte
	var fval []byte
	var fieldSize uint32
	var i uint32
	var err error
	for i = 1; i < uint32(len(raw)); {
		if raw[i] != sep {
			return nil, ErrBadPayload
		}
		i++

		fieldSize = binary.BigEndian.Uint32(raw[i : i+4])
		i += 4

		b = bytes.NewBuffer(raw[i : i+fieldSize])
		r, err = zlib.NewReader(b)
		if err != nil {
			return nil, err
		}

		_, fname, err = util.ReadBody(r, int(fieldSize))
		if err != nil {
			return nil, err
		}
		r.Close()
		i += fieldSize

		fmt.Printf("fs %d cts %v\n", fieldSize, fname)

		if raw[i] != sep {
			return nil, ErrBadPayload
		}
		i++

		fieldSize = binary.BigEndian.Uint32(raw[i : i+4])
		i += 4

		b = bytes.NewBuffer(raw[i : i+fieldSize])
		r, err = zlib.NewReader(b)
		if err != nil {
			return nil, err
		}

		_, fval, err = util.ReadBody(r, int(fieldSize))
		if err != nil {
			return nil, err
		}
		r.Close()
		i += fieldSize

		fmt.Printf("fs %d cts %v\n", fieldSize, fval)

		p.unmarshal(string(fname), fval)

		fname = nil
		fval = nil
	}

	fmt.Printf("unmarshal %d %v %+v\n", v, sep, p)

	return &p, nil
}
