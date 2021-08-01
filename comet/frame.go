package comet

import (
	"bytes"
	"encoding"
	"encoding/binary"
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

type Frame interface {
	OpCode() OpCode
	SetOpCode(OpCode)
	Payload() []byte
	SetPayload([]byte)
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

func IsValidOpCode(code OpCode) bool {

	return true
}

type TCPFrame struct {
	opCode  OpCode
	payload []byte
}

func NewTCPFrame(opCode OpCode, payload []byte) Frame {
	return &TCPFrame{opCode, payload}
}

func (f *TCPFrame) OpCode() OpCode {
	return f.opCode
}

func (f *TCPFrame) SetOpCode(opCode OpCode) {
	f.opCode = opCode
}

func (f *TCPFrame) Payload() []byte {
	res := make([]byte, len(f.payload))
	copy(res, f.payload)
	return res
}

func (f *TCPFrame) SetPayload(payload []byte) {
	f.payload = make([]byte, len(payload))
	copy(f.payload, payload)
}

func (f *TCPFrame) MarshalBinary() (data []byte, err error) {
	// byte[]{OpCode, len(payload), payload}
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, byte(f.opCode)); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.BigEndian, int32(len(f.payload))); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.BigEndian, f.payload); err != nil {
		return nil, err
	}
	//fmt.Println("MarshalBinary", buf.Bytes())
	return buf.Bytes(), nil
}

func (f *TCPFrame) UnmarshalBinary(data []byte) error {
	//fmt.Println("UnmarshalBinary", data)
	r := bytes.NewReader(data)
	if err := binary.Read(r, binary.BigEndian, &f.opCode); err != nil {
		return err
	}
	var length int32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return err
	}
	payload := make([]byte, length)
	if err := binary.Read(r, binary.BigEndian, payload); err != nil {
		return err
	}
	f.payload = payload
	return nil
}

/*
func (f *TCPFrame) GobEncode() ([]byte, error) {
	// byte[]{OpCode, len(payload), payload}
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, byte(f.opCode)); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.BigEndian, int32(len(f.payload))); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.BigEndian, f.payload); err != nil {
		return nil, err
	}
	fmt.Println("GobEncode", buf.Bytes())
	return buf.Bytes(), nil
}

func (f *TCPFrame) GobDecode(b []byte) error {
	fmt.Println("GobDecode", b)
	r := bytes.NewReader(b)
	if err := binary.Read(r, binary.BigEndian, &f.opCode); err != nil {
		return err
	}
	var length int32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return err
	}
	payload := make([]byte, length)
	if err := binary.Read(r, binary.BigEndian, payload); err != nil {
		return err
	}
	f.payload = payload
	return nil
}
*/
