package comet

import (
	"encoding"
	"errors"
)

type OpCode byte

// Opcode type
const (
	OpContinuation OpCode = 0x0
	OpText         OpCode = 0x1
	OpBinary       OpCode = 0x2
	OpClose        OpCode = 0x8
	OpPing         OpCode = 0x9
	OpPong         OpCode = 0xa
)

var opCodes = map[OpCode]struct{}{
	OpContinuation: {},
	OpText:         {},
	OpBinary:       {},
	OpClose:        {},
	OpPing:         {},
	OpPong:         {},
}

var ErrInvalidOpCode = errors.New("Invalid OpCode")

func IsValidOpCode(code OpCode) bool {
	_, ok := opCodes[code]
	return ok
}

type Frame interface {
	OpCode() OpCode
	//SetOpCode(OpCode)
	Payload() []byte
	//SetPayload([]byte)
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}
