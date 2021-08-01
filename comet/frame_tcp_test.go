package comet

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPFrame_NewTCPFrame2(t *testing.T) {

	t.Run("Normal", func(t *testing.T) {
		opCode, payload := OpBinary, []byte{1, 2, 3, 4}

		f, err := NewTCPFrame(OpCode(opCode), payload)
		assert.NoError(t, err)

		opCode = OpClose
		payload[0] = 5

		assert.Equal(t, OpBinary, f.OpCode())
		assert.Equal(t, []byte{1, 2, 3, 4}, f.Payload())
	})

	t.Run("Invalid OpCode", func(t *testing.T) {
		opCode, payload := OpCode(0x3), []byte{1, 2, 3, 4}

		f, err := NewTCPFrame(OpCode(opCode), payload)

		assert.Nil(t, f)
		assert.ErrorIs(t, err, ErrInvalidOpCode)
	})

	t.Run("Empty Payload", func(t *testing.T) {
		opCode, payload := OpBinary, []byte{}

		f, err := NewTCPFrame(OpCode(opCode), payload)

		assert.NotNil(t, f)
		assert.NoError(t, err)
		assert.Equal(t, OpBinary, f.OpCode())
		assert.Equal(t, payload, f.Payload())
	})

	t.Run("Nil Payload", func(t *testing.T) {
		opCode, payload := OpBinary, ([]byte)(nil)

		f, err := NewTCPFrame(OpCode(opCode), payload)

		assert.NotNil(t, f)
		assert.NoError(t, err)
		assert.Equal(t, OpBinary, f.OpCode())
		assert.Equal(t, payload, f.Payload())
	})

}

func TestTCPFrame_MarshalBinary_UnmarshalBinary(t *testing.T) {
	var m tcpFrame
	f, _ := NewTCPFrame(OpClose, []byte{0x11, 0x22, 0x33})
	gob.Register(m)

	network := new(bytes.Buffer)

	enc := gob.NewEncoder(network)
	dec := gob.NewDecoder(network)

	err := enc.Encode(f)
	assert.NoError(t, err)

	var q tcpFrame

	err = dec.Decode(&q)
	assert.NoError(t, err)

	assert.Equal(t, f, &q)
}
