package comet

import (
	"encoding/gob"
	"net"
)

type TCPConn struct {
	net.Conn
	encoder *gob.Encoder
	decoder *gob.Decoder
}

func NewConn(conn net.Conn) Conn {
	return &TCPConn{
		Conn:    conn,
		encoder: gob.NewEncoder(conn),
		decoder: gob.NewDecoder(conn),
	}
}

func (t *TCPConn) ReadFrame() (Frame, error) {
	var f tcpFrame
	if err := t.decoder.Decode(&f); err != nil {
		return nil, err
	}
	return &f, nil
}

func (t *TCPConn) WriteFrame(f Frame) error {
	return t.encoder.Encode(f)
}

/*
func (t *TCPConn) Flush() error {
	return nil
}
*/
