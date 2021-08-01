package comet

import "net"

type Conn interface {
	net.Conn
	ReadFrame() (Frame, error)
	// WriteFrame(OpCode, []byte) error
	WriteFrame(Frame) error
	// Flush() error
}
