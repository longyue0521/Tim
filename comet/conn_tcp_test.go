package comet

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type MockTCPConn struct {
	Buffer *bytes.Buffer
}

// Read reads data from the connection.
// Read can be made to time out and return an error after a fixed
// time limit; see SetDeadline and SetReadDeadline.
func (m *MockTCPConn) Read(b []byte) (n int, err error) {
	n, err = m.Buffer.Read(b)
	return
}

// Write writes data to the connection.
// Write can be made to time out and return an error after a fixed
// time limit; see SetDeadline and SetWriteDeadline.
func (m *MockTCPConn) Write(b []byte) (n int, err error) {
	n, err = m.Buffer.Write(b)
	return
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (m *MockTCPConn) Close() error {
	panic("not implemented") // TODO: Implement
}

// LocalAddr returns the local network address.
func (m *MockTCPConn) LocalAddr() net.Addr {
	panic("not implemented") // TODO: Implement
}

// RemoteAddr returns the remote network address.
func (m *MockTCPConn) RemoteAddr() net.Addr {
	panic("not implemented") // TODO: Implement
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail instead of blocking. The deadline applies to all future
// and pending I/O, not just the immediately following call to
// Read or Write. After a deadline has been exceeded, the
// connection can be refreshed by setting a deadline in the future.
//
// If the deadline is exceeded a call to Read or Write or to other
// I/O methods will return an error that wraps os.ErrDeadlineExceeded.
// This can be tested using errors.Is(err, os.ErrDeadlineExceeded).
// The error's Timeout method will return true, but note that there
// are other possible errors for which the Timeout method will
// return true even if the deadline has not been exceeded.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (m *MockTCPConn) SetDeadline(t time.Time) error {
	panic("not implemented") // TODO: Implement
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (m *MockTCPConn) SetReadDeadline(t time.Time) error {
	panic("not implemented") // TODO: Implement
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (m *MockTCPConn) SetWriteDeadline(t time.Time) error {
	panic("not implemented") // TODO: Implement
}

func TestTCPConn_ReadFrame2(t *testing.T) {

	t.Run("Normal", func(t *testing.T) {
		conn := NewConn(&MockTCPConn{Buffer: new(bytes.Buffer)})
		inputFrames := []Frame{
			NewTCPFrame(OpBinary, []byte("Hello")),
			NewTCPFrame(OpClose, []byte("World")),
		}

		for _, frame := range inputFrames {
			err := conn.WriteFrame(frame)
			assert.NoError(t, err)

			f, err := conn.ReadFrame()
			assert.NoError(t, err)

			assert.Equal(t, frame, f)
		}
	})

	t.Run("Dirty Read", func(t *testing.T) {
		conn := NewConn(&MockTCPConn{Buffer: new(bytes.Buffer)})
		// 直接读取
		f, err := conn.ReadFrame()
		assert.ErrorIs(t, err, io.EOF)
		assert.Nil(t, f)
	})

	t.Run("Not Matched Read", func(t *testing.T) {
		conn := NewConn(&MockTCPConn{Buffer: new(bytes.Buffer)})

		// 写入一个
		frame := NewTCPFrame(OpBinary, []byte("Hello"))
		err := conn.WriteFrame(frame)
		assert.NoError(t, err)

		// 再读取一个
		f, err := conn.ReadFrame()
		assert.NoError(t, err)

		// 匹配
		assert.Equal(t, frame, f)

		// 再读一次，不匹配
		f, err = conn.ReadFrame()
		assert.ErrorIs(t, err, io.EOF)
		assert.Nil(t, f)
	})
}
