package comet

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"net"

	"github.com/gobwas/ws"
)

type Conn interface {
	net.Conn
	ReadFrame() (Frame, error)
	WriteFrame(Frame) error
	Flush() error
}

var (
	_ Conn = &tcpConn{}
	_ Conn = &wsServerConn{}
	_ Conn = &wsClientConn{}
)

type tcpConn struct {
	net.Conn
	encoder *gob.Encoder
	decoder *gob.Decoder
	endian  binary.ByteOrder
}

func NewTCPConn(conn net.Conn) Conn {
	return &tcpConn{
		Conn:    conn,
		encoder: gob.NewEncoder(conn),
		decoder: gob.NewDecoder(conn),
		endian:  binary.BigEndian,
	}
}

func (t *tcpConn) ReadFrame() (Frame, error) {
	f := tcpFrameUnmarshaler{Frame{}, t.endian}
	if err := t.decoder.Decode(&f); err != nil {
		return Frame{}, err
	}
	return f.Frame, nil
}

func (t *tcpConn) WriteFrame(f Frame) error {
	frame := f
	fm := tcpFrameMarshaler{frame, t.endian}
	return t.encoder.Encode(&fm)
}

func (t *tcpConn) Flush() error {
	return nil
}

type tcpFrameMarshaler struct {
	Frame
	endian binary.ByteOrder
}

func (f *tcpFrameMarshaler) MarshalBinary() (data []byte, err error) {
	// data = byte[]{byte(OpCode), int32(len(payload)), payload}
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, f.endian, byte(f.OpCode)); err != nil {
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
	if err := binary.Read(r, f.endian, &f.OpCode); err != nil {
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

type wsServerConn struct {
	net.Conn
}

func NewWSConn(conn net.Conn, isClientSide bool) (Conn, error) {
	wsConn := &wsServerConn{conn}
	_, err := ws.Upgrade(wsConn.Conn)
	if err != nil {
		return nil, err
	}
	if isClientSide {
		return &wsClientConn{wsConn}, nil
	}
	return wsConn, nil
}

func (w *wsServerConn) ReadFrame() (Frame, error) {
	f, err := ws.ReadFrame(w.Conn)
	if err != nil {
		return Frame{}, err
	}
	if f.Header.Masked {
		ws.Cipher(f.Payload, f.Header.Mask, 0)
	}
	return Frame{OpCode: OpCode(f.Header.OpCode), Payload: f.Payload}, nil
}

// WriteFrame(OpCode, []byte) error
func (w *wsServerConn) WriteFrame(f Frame) error {
	frame := ws.NewFrame(ws.OpCode(f.OpCode), true, f.Payload)
	return ws.WriteFrame(w.Conn, frame)
}

func (w *wsServerConn) Flush() error {
	return nil
}

type wsClientConn struct {
	*wsServerConn
}

func (w *wsClientConn) WriteFrame(f Frame) error {
	frame := ws.NewFrame(ws.OpCode(f.OpCode), true, f.Payload)
	frame = ws.MaskFrameInPlaceWith(frame, frame.Header.Mask)
	return ws.WriteFrame(w.Conn, frame)
}
