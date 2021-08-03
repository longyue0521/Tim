package comet

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"net"
)

type tcpConn struct {
	net.Conn
	encoder *gob.Encoder
	decoder *gob.Decoder
}

func NewTCPConn(conn net.Conn) Conn {
	return &tcpConn{
		Conn:    conn,
		encoder: gob.NewEncoder(conn),
		decoder: gob.NewDecoder(conn),
	}
}

func (t *tcpConn) ReadFrame() (Frame, error) {
	f := frameUnmarshaler{Frame{}}
	if err := t.decoder.Decode(&f); err != nil {
		return Frame{}, err
	}
	return f.Frame, nil
}

func (t *tcpConn) WriteFrame(f Frame) error {
	frame := f
	fm := frameMarshaler{frame}
	return t.encoder.Encode(&fm)
}

func (t *tcpConn) Flush() error {
	return nil
}

type frameMarshaler struct {
	Frame
}

func (f *frameMarshaler) MarshalBinary() (data []byte, err error) {
	// data = byte[]{byte(OpCode), int32(len(payload)), payload}
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, byte(f.OpCode)); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.BigEndian, int32(len(f.Payload))); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.BigEndian, f.Payload); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type frameUnmarshaler struct {
	Frame
}

func (f *frameUnmarshaler) UnmarshalBinary(data []byte) error {
	// data = byte[]{byte(OpCode), int32(len(payload)), payload}
	r := bytes.NewReader(data)
	if err := binary.Read(r, binary.BigEndian, &f.OpCode); err != nil {
		return err
	}

	var length int32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return err
	}
	if length > 0 {
		payload := make([]byte, length)
		if err := binary.Read(r, binary.BigEndian, payload); err != nil {
			return err
		}
		f.Payload = payload
	}

	return nil
}
