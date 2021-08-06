package conn

import (
	"net"
)

type Conn interface {
	net.Conn
	ReadFrame() (Frame, error)
	WriteFrame(Frame) error
	Flush() error
}
