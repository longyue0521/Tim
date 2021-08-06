package web

import (
	"bufio"
	"context"
	"net"

	"github.com/gobwas/ws"
	. "github.com/longyue0521/Tim/comet/conn"
)

var (
	_ Conn = &wsServerConn{}
	_ Conn = &wsClientConn{}
)

func NewServerConn(conn net.Conn) (Conn, error) {
	wsConn := &wsServerConn{conn}
	_, err := ws.Upgrade(wsConn.Conn)
	if err != nil {
		// TODO: redefine error
		return nil, err
	}
	return wsConn, nil
}

type wsServerConn struct {
	net.Conn
}

func (w *wsServerConn) ReadFrame() (Frame, error) {
	f, err := ws.ReadFrame(w.Conn)
	if err != nil {
		return Frame{}, err
	}
	f = ws.UnmaskFrameInPlace(f)
	return Frame{Opcode: OpCode(f.Header.OpCode), Payload: f.Payload}, nil
}

func (w *wsServerConn) WriteFrame(f Frame) error {
	frame := ws.NewFrame(ws.OpCode(f.Opcode), true, f.Payload)
	return ws.WriteFrame(w.Conn, frame)
}

func (w *wsServerConn) Flush() error {
	return nil
}

func NewClientConn(address string) (Conn, error) {
	dialer := ws.Dialer{}
	conn, buf, _, err := dialer.Dial(context.Background(), address)
	if err != nil {
		// TODO: redefine error
		return nil, err
	}
	return &wsClientConn{conn, buf}, err
}

type wsClientConn struct {
	net.Conn
	buf *bufio.Reader
}

func (w *wsClientConn) ReadFrame() (Frame, error) {
	var f ws.Frame
	var err error
	if w.buf != nil {
		f, err = ws.ReadFrame(w.buf)
	} else {
		f, err = ws.ReadFrame(w.Conn)
	}
	if err != nil {
		return Frame{}, err
	}
	return Frame{Opcode: OpCode(f.Header.OpCode), Payload: f.Payload}, nil
}

func (w *wsClientConn) WriteFrame(f Frame) error {
	frame := ws.NewFrame(ws.OpCode(f.Opcode), true, f.Payload)
	frame = ws.MaskFrameInPlaceWith(frame, frame.Header.Mask)
	return ws.WriteFrame(w.Conn, frame)
}

func (w *wsClientConn) Flush() error {
	return nil
}
