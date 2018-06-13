package etfapi

import (
	"reflect"
	"testing"

	"github.com/gsmcwhirter/eso-discord/pkg/discordapi/constants"
)

func TestPayload_Marshal(t *testing.T) {
	var one = 1
	tests := []struct {
		name    string
		p       Payload
		want    []byte
		wantErr bool
	}{
		{
			name: "base case",
			p: Payload{
				OpCode: 10,
				Data: map[string]Element{
					"_trace": Element{
						Code: 108,
						Val:  nil,
						Vals: []Element{
							Element{
								Code: 109,
								Val:  []byte{103, 97, 116, 101, 119, 97, 121, 45, 112, 114, 100, 45, 109, 97, 105, 110, 45, 118, 109, 116, 107},
								Vals: nil,
							},
						},
					},
					"heartbeat_interval": Element{
						Code: 98,
						Val:  []byte{0, 0, 161, 34},
						Vals: nil,
					},
				},
				SeqNum:    &one,
				EventName: nil,
			},
			want: []byte{
				131,             // start code
				116, 0, 0, 0, 3, // map 1 length 3

				100, 0, 1, 100, // map 1[0] key [d]
				116, 0, 0, 0, 2, // map1[0] val (map 2 length 2)

				100, 0, 6, 95, 116, 114, 97, 99, 101, // map 2[0] key
				108, 0, 0, 0, 1, // map 2[0] val (list length 1)
				109, 0, 0, 0, 21, 103, 97, 116, 101, 119, 97, 121, 45, 112, 114, 100, 45, 109, 97, 105, 110, 45, 118, 109, 116, 107, 106, // list entry binary

				100, 0, 18, 104, 101, 97, 114, 116, 98, 101, 97, 116, 95, 105, 110, 116, 101, 114, 118, 97, 108, // map 2[1] key
				98, 0, 0, 161, 34, // map 2[1] val

				100, 0, 2, 111, 112, // map 1[1] key [op]
				97, 10, // map 1[1] val

				100, 0, 1, 115, // map 1[2] key [s]
				98, 0, 0, 0, 1, // map 1[2] val
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.p.Marshal()
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Payload.Marshal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	type args struct {
		raw []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *Payload
		wantErr bool
	}{
		{
			name: "base case",
			args: args{[]byte{
				131,             // start code
				116, 0, 0, 0, 4, // map 1 length 4

				100, 0, 1, 100, // map 1[0] key
				116, 0, 0, 0, 2, // map1[0] val (map 2 length 2)

				100, 0, 6, 95, 116, 114, 97, 99, 101, // map 2[0] key
				108, 0, 0, 0, 1, // map 2[0] val (list length 1)
				109, 0, 0, 0, 21, 103, 97, 116, 101, 119, 97, 121, 45, 112, 114, 100, 45, 109, 97, 105, 110, 45, 118, 109, 116, 107, 106, // list entry binary

				100, 0, 18, 104, 101, 97, 114, 116, 98, 101, 97, 116, 95, 105, 110, 116, 101, 114, 118, 97, 108, // map 2[1] key
				98, 0, 0, 161, 34, // map 2[1] val

				100, 0, 2, 111, 112, // map 1[1] key
				97, 10, // map 1[1] val

				100, 0, 1, 115, // map 1[2] key
				100, 0, 3, 110, 105, 108, // map 1[2] val

				100, 0, 1, 116, // map 1[3] key
				100, 0, 3, 110, 105, 108, // map 1[3] val
			}},
			want: &Payload{
				OpCode: 10,
				Data: map[string]Element{
					"_trace": Element{
						Code: 108,
						Val:  nil,
						Vals: []Element{
							Element{
								Code: 109,
								Val:  []byte{103, 97, 116, 101, 119, 97, 121, 45, 112, 114, 100, 45, 109, 97, 105, 110, 45, 118, 109, 116, 107},
								Vals: nil,
							},
						},
					},
					"heartbeat_interval": Element{
						Code: 98,
						Val:  []byte{0, 0, 161, 34},
						Vals: nil,
					},
				},
				SeqNum:    nil,
				EventName: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Unmarshal(tt.args.raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unmarshal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewElement(t *testing.T) {
	type args struct {
		code ETFCode
		val  interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    *Element
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewElement(tt.args.code, tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewElement() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewElement() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPayload_String(t *testing.T) {
	type fields struct {
		OpCode    constants.OpCode
		Data      map[string]Element
		SeqNum    *int
		EventName *string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Payload{
				OpCode:    tt.fields.OpCode,
				Data:      tt.fields.Data,
				SeqNum:    tt.fields.SeqNum,
				EventName: tt.fields.EventName,
			}
			if got := p.String(); got != tt.want {
				t.Errorf("Payload.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestPayload_unmarshal(t *testing.T) {
// 	type args struct {
// 		key string
// 		val Element
// 	}
// 	tests := []struct {
// 		name    string
// 		p       *Payload
// 		args    args
// 		wantErr bool
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.p.unmarshal(tt.args.key, tt.args.val); (err != nil) != tt.wantErr {
// 				t.Errorf("Payload.unmarshal() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func Test_writeAtom(t *testing.T) {
// 	type args struct {
// 		b   *bytes.Buffer
// 		val string
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		wantErr bool
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := writeAtom(tt.args.b, tt.args.val); (err != nil) != tt.wantErr {
// 				t.Errorf("writeAtom() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func Test_writeMapListLength(t *testing.T) {
// 	type args struct {
// 		b *bytes.Buffer
// 		n int
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		wantErr bool
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := writeMapListLength(tt.args.b, tt.args.n); (err != nil) != tt.wantErr {
// 				t.Errorf("writeMapListLength() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func Test_writeElement(t *testing.T) {
// 	type args struct {
// 		b *bytes.Buffer
// 		e Element
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		wantErr bool
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := writeElement(tt.args.b, tt.args.e); (err != nil) != tt.wantErr {
// 				t.Errorf("writeElement() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func Test_marshalInterface(t *testing.T) {
// 	type args struct {
// 		code ETFCode
// 		val  interface{}
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    []byte
// 		wantErr bool
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := marshalInterface(tt.args.code, tt.args.val)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("marshalInterface() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("marshalInterface() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func Test_unmarshalSlice(t *testing.T) {
// 	type args struct {
// 		raw         []byte
// 		numElements int
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    uint32
// 		want1   []Element
// 		wantErr bool
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, got1, err := unmarshalSlice(tt.args.raw, tt.args.numElements)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("unmarshalSlice() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if got != tt.want {
// 				t.Errorf("unmarshalSlice() got = %v, want %v", got, tt.want)
// 			}
// 			if !reflect.DeepEqual(got1, tt.want1) {
// 				t.Errorf("unmarshalSlice() got1 = %v, want %v", got1, tt.want1)
// 			}
// 		})
// 	}
// }
