package comet

import (
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
	var f tcpFrame
	if err := t.decoder.Decode(&f); err != nil {
		return nil, err
	}
	return &f, nil
}

func (t *tcpConn) WriteFrame(f Frame) error {
	return t.encoder.Encode(f)
}

/*
func (t *TCPConn) Flush() error {
	return nil
}
*/
