package etfapi

import (
	"reflect"
	"testing"
)

func TestPayload_Marshal(t *testing.T) {
	tests := []struct {
		name string
		p    Payload
		want []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Marshal(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Payload.Marshal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPayload_unmarshal(t *testing.T) {
	type args struct {
		key string
		val element
	}
	tests := []struct {
		name    string
		p       *Payload
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.unmarshal(tt.args.key, tt.args.val); (err != nil) != tt.wantErr {
				t.Errorf("Payload.unmarshal() error = %v, wantErr %v", err, tt.wantErr)
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
			args: args{[]byte{131, 116, 0, 0, 0, 4, 100, 0, 1, 100, 116, 0, 0, 0, 2, 100, 0, 6, 95, 116, 114, 97, 99, 101, 108, 0, 0, 0, 1, 109, 0, 0,
				0, 21, 103, 97, 116, 101, 119, 97, 121, 45, 112, 114, 100, 45, 109, 97, 105, 110, 45, 118, 109, 116, 107, 106, 100, 0, 18, 104, 101, 97, 114, 116, 98, 101,
				97, 116, 95, 105, 110, 116, 101, 114, 118, 97, 108, 98, 0, 0, 161, 34, 100, 0, 2, 111, 112, 97, 10, 100, 0, 1, 115, 100, 0, 3, 110, 105, 108, 100, 0, 1,
				116, 100, 0, 3, 110, 105, 108}},
			want:    &Payload{},
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
			t.Errorf("fail")
		})
	}
}
