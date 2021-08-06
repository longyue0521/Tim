package conn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidOpCode(t *testing.T) {
	type args struct {
		code OpCode
	}

	tests := map[string]struct {
		args args
		want bool
	}{
		"valid OpContinuation": {args{OpContinuation}, true},
		"valid OpText":         {args{OpText}, true},
		"valid OpBinary":       {args{OpBinary}, true},
		"valid OpClose":        {args{OpClose}, true},
		"valid OpPing":         {args{OpPing}, true},
		"valid OpPong":         {args{OpPong}, true},

		"invalid 0x3":  {args{OpCode(0x3)}, false},
		"invalid 0xff": {args{OpCode(0xff)}, false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidOpCode(tt.args.code))
		})
	}
}
