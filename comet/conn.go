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
}

func NewTCPConn(conn net.Conn) Conn {
	return &tcpConn{
		Conn:    conn,
		encoder: gob.NewEncoder(conn),
		decoder: gob.NewDecoder(conn),
	}
}

func (t *tcpConn) ReadFrame() (Frame, error) {
	f := tcpFrameUnmarshaler{Frame{}}
	if err := t.decoder.Decode(&f); err != nil {
		return Frame{}, err
	}
	return f.Frame, nil
}

func (t *tcpConn) WriteFrame(f Frame) error {
	frame := f
	fm := tcpFrameMarshaler{frame}
	return t.encoder.Encode(&fm)
}

func (t *tcpConn) Flush() error {
	return nil
}

type tcpFrameMarshaler struct {
	Frame
}

func (f *tcpFrameMarshaler) MarshalBinary() (data []byte, err error) {
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

type tcpFrameUnmarshaler struct {
	Frame
}

func (f *tcpFrameUnmarshaler) UnmarshalBinary(data []byte) error {
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

type wsServerConn struct {
	net.Conn
}

func NewWSConn(conn net.Conn, isClient bool) Conn {
	if isClient {
		return &wsClientConn{&wsServerConn{Conn: conn}}
	}
	return &wsServerConn{conn}
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
