package comet

import (
	"bytes"
	"encoding/binary"
)

type tcpFrame struct {
	opCode  OpCode
	payload []byte
}

func NewTCPFrame(opCode OpCode, payload []byte) (Frame, error) {

	if !IsValidOpCode(opCode) {
		return nil, ErrInvalidOpCode
	}

	if payload == nil {
		return &tcpFrame{opCode: opCode}, nil
	}

	f := &tcpFrame{opCode, payload}
	f.payload = make([]byte, len(payload))
	copy(f.payload, payload)

	return f, nil
}

func (f *tcpFrame) OpCode() OpCode {
	return f.opCode
}

func (f *tcpFrame) Payload() []byte {
	if f.payload == nil {
		return nil
	}
	res := make([]byte, len(f.payload))
	copy(res, f.payload)
	return res
}

func (f *tcpFrame) MarshalBinary() (data []byte, err error) {
	// data = byte[]{byte(OpCode), int32(len(payload)), payload}
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

	return buf.Bytes(), nil
}

func (f *tcpFrame) UnmarshalBinary(data []byte) error {
	// data = byte[]{byte(OpCode), int32(len(payload)), payload}
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
