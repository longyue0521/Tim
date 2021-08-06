package tcp

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"net"

	. "github.com/longyue0521/Tim/comet/conn"
	"github.com/pkg/errors"
)

var (
	_                  Conn = &tcpServerConn{}
	_                  Conn = &tcpClientConn{}
	ErrInvalidArgument      = errors.New("invalid argument")
)

func NewServerConn(conn net.Conn) (Conn, error) {
	if conn == nil {
		return nil, errors.Wrapf(ErrInvalidArgument, "%s conn net.Conn", ErrInvalidArgument)
	}
	return &tcpServerConn{
		Conn:    conn,
		encoder: gob.NewEncoder(conn),
		decoder: gob.NewDecoder(conn),
		endian:  binary.BigEndian,
	}, nil
}

type tcpServerConn struct {
	net.Conn
	encoder *gob.Encoder
	decoder *gob.Decoder
	endian  binary.ByteOrder
}

func (t *tcpServerConn) ReadFrame() (Frame, error) {
	f := tcpFrameUnmarshaler{Frame{}, t.endian}
	if err := t.decoder.Decode(&f); err != nil {
		return Frame{}, err
	}
	return f.Frame, nil
}

func (t *tcpServerConn) WriteFrame(f Frame) error {
	frame := f
	fm := tcpFrameMarshaler{frame, t.endian}
	return t.encoder.Encode(&fm)
}

func (t *tcpServerConn) Flush() error {
	return nil
}

type tcpFrameMarshaler struct {
	Frame
	endian binary.ByteOrder
}

func (f *tcpFrameMarshaler) MarshalBinary() (data []byte, err error) {
	// data = byte[]{byte(OpCode), int32(len(payload)), payload}
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, f.endian, byte(f.Opcode)); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, f.endian, int32(len(f.Payload))); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, f.endian, f.Payload); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type tcpFrameUnmarshaler struct {
	Frame
	endian binary.ByteOrder
}

func (f *tcpFrameUnmarshaler) UnmarshalBinary(data []byte) error {
	// data = byte[]{byte(OpCode), int32(len(payload)), payload}
	r := bytes.NewReader(data)
	if err := binary.Read(r, f.endian, &f.Opcode); err != nil {
		return err
	}

	var length int32
	if err := binary.Read(r, f.endian, &length); err != nil {
		return err
	}
	if length <= 0 {
		return nil
	}

	payload := make([]byte, length)
	if err := binary.Read(r, f.endian, payload); err != nil {
		return err
	}
	f.Payload = payload

	return nil
}

func NewClientConn(address string) (Conn, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		// TODO: rewrite error
		return nil, err
	}
	server, err := NewServerConn(conn)
	if err != nil {
		return nil, err
	}
	return &tcpClientConn{server}, nil
}

type tcpClientConn struct {
	Conn
}
