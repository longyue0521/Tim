package comet

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/stretchr/testify/assert"
)

type MockNetConn struct {
	*bytes.Buffer
}

func (m *MockNetConn) Read(b []byte) (n int, err error) {
	return m.Buffer.Read(b)
}

func (m *MockNetConn) Write(b []byte) (n int, err error) {
	return m.Buffer.Write(b)
}

func (m *MockNetConn) Close() error {
	panic("not implemented") // TODO: Implement
}

func (m *MockNetConn) LocalAddr() net.Addr {
	panic("not implemented") // TODO: Implement
}

func (m *MockNetConn) RemoteAddr() net.Addr {
	panic("not implemented") // TODO: Implement
}

func (m *MockNetConn) SetDeadline(t time.Time) error {
	panic("not implemented") // TODO: Implement
}

func (m *MockNetConn) SetReadDeadline(t time.Time) error {
	panic("not implemented") // TODO: Implement
}

func (m *MockNetConn) SetWriteDeadline(t time.Time) error {
	panic("not implemented") // TODO: Implement
}
func TestConn(t *testing.T) {
	tcpConn := &MockNetConn{&bytes.Buffer{}}

	frames := []Frame{
		{},
		{OpPing, nil},
		{OpBinary, []byte("Hello")},
		{OpClose, []byte("World")},
	}

	tests := map[string]struct {
		client Conn
		server Conn
	}{
		"TCP": {
			client: NewTCPConn(tcpConn),
			server: NewTCPConn(tcpConn),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {

			MatchedWriteAndRead := func(client Conn, server Conn) {
				for _, frame := range frames {
					err := client.WriteFrame(frame)
					assert.NoError(t, err)

					f, err := server.ReadFrame()
					assert.NoError(t, err)

					assert.Equal(t, frame, f)
				}
			}

			MatchedWriteAndRead(tt.client, tt.server)
			MatchedWriteAndRead(tt.server, tt.client)

			MatchedMultiWriteAndMultiRead := func(client Conn, server Conn) {
				for _, frame := range frames {
					server.WriteFrame(frame)
				}
				for _, frame := range frames {
					f, err := client.ReadFrame()
					assert.Equal(t, frame, f)
					assert.NoError(t, err)
				}
			}

			MatchedMultiWriteAndMultiRead(tt.client, tt.server)
			MatchedMultiWriteAndMultiRead(tt.server, tt.client)
		})
	}
}

func TestEmptyPayload(t *testing.T) {
	tcpConn := &MockNetConn{&bytes.Buffer{}}

	tests := map[string]struct {
		client Conn
		server Conn
	}{
		"TCP": {
			client: NewTCPConn(tcpConn),
			server: NewTCPConn(tcpConn),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			WriteAndReadEmptyPayload := func(client, server Conn) {
				// 写入一个
				frame := Frame{OpBinary, []byte{}}
				err := server.WriteFrame(frame)
				assert.NoError(t, err)

				// 读取一个
				f, err := client.ReadFrame()
				assert.NoError(t, err)

				// 匹配
				// 注意：这里读取出来的是Frame{OpBinary, nil}
				//      而不是Frame{OpBinary, []byte{}}
				//      下面的断言会出错
				//assert.Equal(t, frame, f)
				assert.Equal(t, OpBinary, f.OpCode)
				assert.Nil(t, f.Payload)
			}

			WriteAndReadEmptyPayload(tt.client, tt.server)
			WriteAndReadEmptyPayload(tt.server, tt.client)
		})
	}
}

func TestDirtyRead(t *testing.T) {
	tcpConn := &MockNetConn{&bytes.Buffer{}}

	tests := map[string]struct {
		conn Conn
	}{
		"TCP Client": {NewTCPConn(tcpConn)},
		"TCP Server": {NewTCPConn(tcpConn)},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// 直接读取
			f, err := tt.conn.ReadFrame()

			assert.ErrorIs(t, err, io.EOF)
			assert.Equal(t, Frame{}, f)
			assert.Equal(t, OpContinuation, f.OpCode)
			assert.Equal(t, []byte(nil), f.Payload)
			assert.Nil(t, f.Payload)
		})
	}
}

func TestUnmachedRead(t *testing.T) {
	tcpConn := &MockNetConn{&bytes.Buffer{}}

	tests := map[string]struct {
		client Conn
		server Conn
	}{
		"TCP": {
			client: NewTCPConn(tcpConn),
			server: NewTCPConn(tcpConn),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {

			OneWriteTwoRead := func(client, server Conn) {
				// 写入一个
				frame := Frame{OpBinary, []byte("Hello")}
				err := server.WriteFrame(frame)
				assert.NoError(t, err)

				// 读取一个
				f, err := client.ReadFrame()
				assert.NoError(t, err)

				// 匹配
				assert.Equal(t, frame, f)

				// 再读一次，不匹配
				f, err = client.ReadFrame()
				assert.ErrorIs(t, err, io.EOF)
				assert.Equal(t, Frame{}, f)
			}

			OneWriteTwoRead(tt.client, tt.server)
			OneWriteTwoRead(tt.server, tt.client)

		})
	}
}

func TestWebsocket_ServerConn_ReadFrame(t *testing.T) {

	frames := []Frame{
		{OpPing, []byte("Ping....")},
		{OpPong, []byte("Pong....")},
		{OpBinary, []byte("websocket")},
	}

	// 创建监听socket
	ln, err := net.Listen("tcp", "127.0.0.1:")
	assert.NoError(t, err)

	// Mock Write-Only websocket 客户端
	go func() {
		url := fmt.Sprintf("ws://%s/", ln.Addr().String())
		conn, _, _, err := ws.Dial(context.Background(), url)
		assert.NoError(t, err)

		// 写
		for _, frame := range frames {
			err = wsutil.WriteClientMessage(conn, ws.OpCode(frame.OpCode), frame.Payload)
			assert.NoError(t, err)
		}

		err = conn.Close()
		assert.NoError(t, err)
	}()

	// 创建服务端socket
	conn, err := ln.Accept()
	assert.NoError(t, err)
	assert.NoError(t, ln.Close())

	// 创建服务端Websocket
	wsConn, err := NewWSConn(conn, false)
	assert.NoError(t, err)

	for _, frame := range frames {
		f, err := wsConn.ReadFrame()
		assert.NoError(t, err)
		assert.Equal(t, frame, f)
	}

	assert.NoError(t, wsConn.Close())
}

func TestWebsocket_ServerConn_WriteFrame(t *testing.T) {

	frames := []Frame{
		{OpPing, []byte("Ping....")},
		{OpPong, []byte("Pong....")},
		{OpBinary, []byte("websocket")},
	}

	// 创建监听socket
	ln, err := net.Listen("tcp", "127.0.0.1:")
	assert.NoError(t, err)

	// Mock Read-Only websocket 客户端
	go func() {

		url := fmt.Sprintf("ws://%s/", ln.Addr().String())
		conn, buf, _, err := ws.Dial(context.Background(), url)
		assert.NotNil(t, buf)
		assert.NoError(t, err)

		// 读
		for _, frame := range frames {
			f, err := ws.ReadFrame(buf)
			assert.NoError(t, err)
			assert.Equal(t, frame, Frame{OpCode: OpCode(f.Header.OpCode), Payload: f.Payload})
		}

		err = conn.Close()
		assert.NoError(t, err)
	}()

	// 创建服务端socket
	conn, err := ln.Accept()
	assert.NoError(t, err)
	assert.NoError(t, ln.Close())

	// 创建服务端Websocket
	wsConn, err := NewWSConn(conn, false)
	assert.NoError(t, err)

	for _, frame := range frames {
		err = wsConn.WriteFrame(frame)
		assert.NoError(t, err)
	}

	assert.NoError(t, wsConn.Close())
}
