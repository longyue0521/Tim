package comet

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPFrame(t *testing.T) {
	var f TCPFrame
	code, payload := OpBinary, []byte{1, 2, 3, 4}
	f.SetOpCode(OpCode(code))
	f.SetPayload(payload)
	code = OpClose
	payload[0] = 2

	assert.Equal(t, []byte{1, 2, 3, 4}, f.Payload())
	assert.Equal(t, OpBinary, f.OpCode())
}

func TestTCPFrame_MarshalBinary_UnmarshalBinary(t *testing.T) {
	var m TCPFrame
	f := NewTCPFrame(OpClose, []byte{0x11, 0x22, 0x33})
	gob.Register(m)

	network := new(bytes.Buffer)
	fmt.Println(network.Bytes())

	enc := gob.NewEncoder(network)
	dec := gob.NewDecoder(network)

	err := enc.Encode(f)
	assert.NoError(t, err)

	fmt.Println(network.Bytes())
	fmt.Println(network.Len())
	var q TCPFrame

	err = dec.Decode(&q)
	assert.NoError(t, err)

	assert.Equal(t, f, &q)
}
