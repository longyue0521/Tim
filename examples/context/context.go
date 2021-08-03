package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

type Channel struct {
	id         string
	ch         chan int
	ctx        context.Context
	cancelFunc context.CancelFunc
	once       sync.Once
}

func (c *Channel) WriteLoop() error {
	for {
		select {
		case n, ok := <-c.ch:
			if !ok {
				continue
			}
			fmt.Println(n)

			for i := 0; i < len(c.ch); i++ {
				fmt.Println(<-c.ch)
			}

		case <-c.ctx.Done():
			return nil
		}
	}
}

func (c *Channel) Push(i int) error {
	// if c.ctx.Err() != nil {
	// 	return fmt.Errorf("channel [%s] closed", c.id)
	// }
	select {
	case <-c.ctx.Done():
		return fmt.Errorf("channel [%s] closed", c.id)
	default:
		c.ch <- i
		return nil
	}

}

func (c *Channel) Close() error {
	c.once.Do(func() {
		close(c.ch)
		c.cancelFunc()
	})
	return nil
}

func main() {
	ch := &Channel{
		ch: make(chan int, 5),
		id: "1A",
	}
	ch.ctx, ch.cancelFunc = context.WithCancel(context.Background())
	//waitGroup(ch)
	errorGroup(ch)
}

func errorGroup(ch *Channel) {
	fmt.Println("errorGroup")
	eg := new(errgroup.Group)

	eg.Go(func() error {
		return ch.WriteLoop()
	})

	eg.Go(func() error {
		for i := 0; i < 100; i++ {
			if err := ch.Push(i); err != nil {
				return fmt.Errorf("Push Error : %s\n", err)
			}
		}
		return nil
	})

	eg.Go(func() error {
		time.Sleep(time.Second * 1)
		return ch.Close()
	})

	if err := eg.Wait(); err != nil {
		fmt.Println(err)
	}
}

func waitGroup(ch *Channel) {
	fmt.Println("WaitGroup")
	var wg sync.WaitGroup

	wg.Add(3)

	go func() {
		defer wg.Done()
		if err := ch.WriteLoop(); err != nil {
			fmt.Printf("Loop Error %s\n", err)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			if err := ch.Push(i); err != nil {
				fmt.Printf("Push Error %s\n", err)
			}
		}
	}()

	go func() {
		defer wg.Done()
		time.Sleep(time.Second * 1)
		ch.Close()
	}()

	wg.Wait()
}
