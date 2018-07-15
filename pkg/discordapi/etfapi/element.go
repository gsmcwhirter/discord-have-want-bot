package etfapi

import (
	"bytes"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

// Element TODOC
type Element struct {
	Code ETFCode
	Val  []byte
	Vals []Element
}

// NewElement TODOC
func NewElement(code ETFCode, val interface{}) (e Element, err error) {
	e.Code = code

	if v, ok := val.([]Element); ok {
		if !code.IsCollection() {
			err = ErrBadElementData
			return
		}
		e.Vals = v

		return
	}

	if v, ok := val.([]byte); ok {
		if code.IsCollection() {
			err = ErrBadElementData
			return
		}
		e.Val = v

		return
	}

	return
}

// NewNilElement TODOC
func NewNilElement() (e Element, err error) {
	e, err = NewAtomElement("nil")
	err = errors.Wrap(err, "could not create Nil Element")
	return
}

// NewBoolElement TODOC
func NewBoolElement(val bool) (e Element, err error) {
	if val {
		e, err = NewAtomElement("true")
	} else {
		e, err = NewAtomElement("false")
	}

	err = errors.Wrap(err, "could not create Bool Element")
	return
}

// NewInt8Element TODOC
func NewInt8Element(val int) (e Element, err error) {
	var v []byte
	v, err = IntToInt8Slice(val)
	if err != nil {
		err = errors.Wrap(err, "could not convert to int8 slice")
		return
	}

	e, err = NewElement(Int8, v)
	err = errors.Wrap(err, "could not create Int8 Element")
	return
}

// NewInt32Element TODOC
func NewInt32Element(val int) (e Element, err error) {
	var v []byte
	v, err = IntToInt32Slice(val)
	if err != nil {
		err = errors.Wrap(err, "could not convert to int32 slice")
		return
	}

	e, err = NewElement(Int32, v)
	err = errors.Wrap(err, "could not create Int32 Element")
	return
}

// NewBinaryElement TODOC
func NewBinaryElement(val []byte) (e Element, err error) {
	e, err = NewElement(Binary, val)
	err = errors.Wrap(err, "could not create binary Element")
	return
}

// NewAtomElement TODOC
func NewAtomElement(val string) (e Element, err error) {
	e, err = NewElement(Atom, []byte(val))
	err = errors.Wrap(err, "could not create atom Element")
	return
}

// NewStringElement TODOC
func NewStringElement(val string) (e Element, err error) {
	e, err = NewElement(Binary, []byte(val))
	err = errors.Wrap(err, "could not create string Element")
	return
}

// NewMapElement TODOC
func NewMapElement(val map[string]Element) (e Element, err error) {
	e2, err := ElementMapToElementSlice(val)
	if err != nil {
		err = errors.Wrap(err, "could not create element slice")
		return
	}

	e, err = NewElement(Map, e2)
	err = errors.Wrap(err, "could not create map Element")
	return
}

// NewListElement TODOC
func NewListElement(val []Element) (e Element, err error) {
	e, err = NewElement(List, val)
	err = errors.Wrap(err, "could not create list Element")
	return
}

// String TODOC
func (e Element) String() string {
	switch e.Code {
	case Map:
		fallthrough
	case List:
		return fmt.Sprintf("Element{Code: %v, Vals: %v}", e.Code, e.Vals)
	default:
		return fmt.Sprintf("Element{Code: %v, Val: %v}", e.Code, e.Val)
	}
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

// PrettyString TODOC
func (e *Element) PrettyString(indent string, skipFirstIndent bool) string {
	b := bytes.Buffer{}

	if e.Code == Binary {
		if skipFirstIndent {
			indent = ""
		}
		_, _ = b.WriteString(fmt.Sprintf("%s%s", indent, string(e.Val)))
		return b.String()
	}

	if skipFirstIndent {
		_, _ = b.WriteString("Element{\n")
	} else {
		_, _ = b.WriteString(fmt.Sprintf("%sElement{\n", indent))
	}

	_, _ = b.WriteString(fmt.Sprintf("%s  Type: %v\n", indent, e.Code))
	if e.Code == List {
		_, _ = b.WriteString(fmt.Sprintf("%s  Vals: [\n", indent))
		for _, v := range e.Vals {
			_, _ = b.WriteString(v.PrettyString(indent+"     ", false))
			_, _ = b.WriteString("\n")
		}
		_, _ = b.WriteString(fmt.Sprintf("%s  ]", indent))
	} else if e.Code == Map {
		_, _ = b.WriteString(fmt.Sprintf("%s  Vals: {\n", indent))
		for i := 0; i < len(e.Vals); i += 2 {
			_, _ = b.WriteString(e.Vals[i].PrettyString(indent+"     ", false))
			_, _ = b.WriteString(": ")
			_, _ = b.WriteString(e.Vals[i+1].PrettyString(indent+"     ", true))
			_, _ = b.WriteString("\n")
		}
		_, _ = b.WriteString(fmt.Sprintf("%s  }", indent))
	} else {
		_, _ = b.WriteString(fmt.Sprintf("%s  Val: %v", indent, e.Val))
	}

	return b.String()
}
