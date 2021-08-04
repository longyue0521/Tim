package comet

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Agent is interface of client side
type Agent interface {
	ID() string
	Push([]byte) error
}

// MessageListener 监听消息
type MessageListener interface {
	// 收到消息回调
	Receive(Agent, []byte)
}

type Channel interface {
	Conn
	Agent
	ReadLoop(lst MessageListener) error
	// SetWriteWait 设置写超时
	SetReadTimeout(time.Duration)
	SetWriteTimeout(time.Duration)
}

const (
	DefaultReadTimeout  = time.Minute * 3
	DefaultWriteTimeout = time.Second * 10
)

type channel struct {
	Conn
	id          string
	payloadChan chan []byte
	m           sync.Mutex
	once        sync.Once
	rdTimeout   time.Duration
	wtTimeout   time.Duration
	ctx         context.Context
	cancelFunc  context.CancelFunc
	//closed *Event
}

func NewChannel(id string, conn Conn) Channel {
	// TODO: add log info
	/*
		log := logger.WithFields(logger.Fields{
			"module": "channel",
			"id":     id,
		})
	*/
	c := &channel{
		id:          id,
		Conn:        conn,
		payloadChan: make(chan []byte, 5),
		//closed: NewEvent(),
		rdTimeout: DefaultReadTimeout,
		wtTimeout: DefaultWriteTimeout,
	}

	c.ctx, c.cancelFunc = context.WithCancel(context.Background())

	return c
}

func (c *channel) writeLoop() error {
	for {
		select {
		case payload := <-c.payloadChan:
			//
			err := c.WriteFrame(Frame{OpBinary, payload})
			if err != nil {
				return err
			}
			// retrive more payload as possible
			for i := 0; i < len(c.payloadChan); i++ {
				err = c.WriteFrame(Frame{OpBinary, <-c.payloadChan})
				if err != nil {
					return err
				}
			}
			//
			err = c.Flush()
			if err != nil {
				return err
			}

		// case <- c.closed.Done():
		// 	return nil
		case <-c.ctx.Done():
			return nil
		}
	}
}

func (c *channel) ID() string { return c.id }

func (c *channel) Push(payload []byte) error {
	// if ch.closed.HasFired() {
	// 	return fmt.Errorf("channel %s has closed", ch.id)
	// }

	select {
	case <-c.ctx.Done():
		return fmt.Errorf("channel %s has closed", c.id)
	default:
		// 异步写
		c.payloadChan <- payload
		return nil
	}

}

func (c *channel) ReadLoop(lst MessageListener) error {
	c.m.Lock()
	defer c.m.Unlock()

	// TODO: add logger

	for {

		frame, err := c.ReadFrame()
		if err != nil {
			return err
		}

		// handle Close Frame
		if frame.OpCode == OpClose {
			// 这里应该主动关闭吧，而不是抛出错误
			// 答：通过抛出错误来提醒调用者停止
			return errors.New("remote close channel")
		}
		// handle Ping Frame
		if frame.OpCode == OpPing {
			// TODO: add log.Trace
			c.WriteFrame(Frame{OpCode: OpPong})
			continue
		}

		payload := frame.Payload
		if len(payload) == 0 {
			continue
		}

		// TODO:
		go lst.Receive(c, payload)
	}
}

func (c *channel) ReadFrame() (Frame, error) {
	// error 问题， 如果失败怎么办？
	_ = c.Conn.SetReadDeadline(time.Now().Add(c.rdTimeout))
	return c.Conn.ReadFrame()
}

func (c *channel) WriteFrame(f Frame) error {
	// error 问题， 如果失败怎么办？
	c.Conn.SetWriteDeadline(time.Now().Add(c.wtTimeout))
	return c.Conn.WriteFrame(f)
}

func (c *channel) Close() error {
	c.once.Do(func() {
		close(c.payloadChan)
		//ch.closed.Fire()
		c.cancelFunc()
	})
	return nil
}

func (c *channel) SetReadTimeout(t time.Duration) {
	c.rdTimeout = t
}

func (c *channel) SetWriteTimeout(t time.Duration) {
	c.wtTimeout = t
}

type ChannelPool interface {
	Add(Channel)
	Del(id string)
	Get(id string) (Channel, bool)
	All() []Channel
}

type pool struct {
	mp *sync.Map
}

func NewChannelPool(n int) ChannelPool {
	return &pool{mp: new(sync.Map)}
}

func (p *pool) Add(c Channel) {
	if c.ID() == "" {
		// add logger
	}
	p.mp.Store(c.ID(), c)
}

func (p *pool) Del(id string) {
	p.mp.Delete(id)
}

func (p *pool) Get(id string) (Channel, bool) {
	if id == "" {
		// add logger
	}
	v, ok := p.mp.Load(id)
	if !ok {
		return nil, false
	}
	return v.(Channel), true
}

func (p *pool) All() []Channel {
	chs := make([]Channel, 0)
	p.mp.Range(func(key interface{}, val interface{}) bool {
		chs = append(chs, val.(Channel))
		return true
	})
	return chs
}
