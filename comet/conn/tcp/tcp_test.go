package tcp

import (
	"fmt"
	"io"
	"net"
	"testing"

	. "github.com/longyue0521/Tim/comet/conn"
	"github.com/stretchr/testify/assert"
)

func TestNewServerConn_NewClientConn(t *testing.T) {
	server, err := NewServerConn(nil)
	assert.Nil(t, server)
	fmt.Println(err)
	assert.Error(t, err)

	client, err := NewClientConn("")
	assert.Nil(t, client)
	assert.Error(t, err)
}

func TestTCPsocket(t *testing.T) {

	frames := []Frame{
		{Opcode: OpPing, Payload: []byte("Ping....")},
		{Opcode: OpPong, Payload: []byte("Pong....")},
		{Opcode: OpBinary, Payload: []byte("websocket")},
	}

	// 创建监听socket
	ln, err := net.Listen("tcp", "127.0.0.1:")
	assert.NoError(t, err)

	// Mock tcpsocket client
	go func() {

		url := ln.Addr().String()
		client, err := NewClientConn(url)
		assert.NoError(t, err)

		// 连续写
		for _, frame := range frames {
			err := client.WriteFrame(frame)
			assert.NoError(t, err)
		}
		// 连续读
		for _, frame := range frames {
			f, err := client.ReadFrame()
			assert.NoError(t, err)
			assert.Equal(t, frame, f)
		}

		err = client.Close()
		assert.NoError(t, err)
	}()

	// 创建服务端socket
	conn, err := ln.Accept()
	assert.NoError(t, err)
	assert.NoError(t, ln.Close())

	// 创建服务端Websocket
	server, err := NewServerConn(conn)
	assert.NoError(t, err)

	for _, frame := range frames {
		// 读
		f, err := server.ReadFrame()
		assert.NoError(t, err)

		assert.Equal(t, frame, f)
		// 回写
		err = server.WriteFrame(f)
		assert.NoError(t, err)
	}

	assert.NoError(t, server.Close())
}

func TestServerConnReadFrame_ClientConnWriteFrame(t *testing.T) {

	frames := []Frame{
		{Opcode: OpPing, Payload: []byte("Ping....")},
		{Opcode: OpPong, Payload: []byte("Pong....")},
		{Opcode: OpBinary, Payload: []byte("websocket")},
	}

	// 创建监听socket
	ln, err := net.Listen("tcp", "127.0.0.1:")
	assert.NoError(t, err)

	// Mock Write-Only websocket client
	go func() {

		url := ln.Addr().String()
		client, err := NewClientConn(url)
		assert.NoError(t, err)

		// 连续写
		for _, frame := range frames {
			err := client.WriteFrame(frame)
			assert.NoError(t, err)
		}

		err = client.Close()
		assert.NoError(t, err)
	}()

	// 创建服务端socket
	conn, err := ln.Accept()
	assert.NoError(t, err)
	assert.NoError(t, ln.Close())

	// 创建服务端Websocket
	server, err := NewServerConn(conn)
	assert.NoError(t, err)

	// 匹配连续读
	for _, frame := range frames {
		f, err := server.ReadFrame()
		assert.NoError(t, err)
		assert.Equal(t, frame, f)
	}

	// 多读一次
	f, err := server.ReadFrame()
	assert.Equal(t, Frame{}, f)
	assert.ErrorIs(t, err, io.EOF)

	assert.NoError(t, server.Close())
}

func TestServerConnWriteFrame_ClientConnReadFrame(t *testing.T) {

	frames := []Frame{
		{Opcode: OpPing, Payload: []byte("Ping....")},
		{Opcode: OpPong, Payload: []byte("Pong....")},
		{Opcode: OpBinary, Payload: []byte("websocket")},
	}

	// 创建监听socket
	ln, err := net.Listen("tcp", "127.0.0.1:")
	assert.NoError(t, err)

	// Mock Read-Only websocket client
	go func() {

		url := ln.Addr().String()
		client, err := NewClientConn(url)
		assert.NoError(t, err)

		// 匹配读
		for _, frame := range frames {
			f, err := client.ReadFrame()
			assert.Equal(t, frame, f)
			assert.NoError(t, err)
		}
		// 多读一次
		f, err := client.ReadFrame()
		assert.Equal(t, Frame{}, f)
		assert.ErrorIs(t, err, io.EOF)

		err = client.Close()
		assert.NoError(t, err)
	}()

	// 创建服务端socket
	conn, err := ln.Accept()
	assert.NoError(t, err)
	assert.NoError(t, ln.Close())

	// 创建服务端Websocket
	server, err := NewServerConn(conn)
	assert.NoError(t, err)

	// 连续写
	for _, frame := range frames {
		err = server.WriteFrame(frame)
		assert.NoError(t, err)
	}

	assert.NoError(t, server.Close())
}
