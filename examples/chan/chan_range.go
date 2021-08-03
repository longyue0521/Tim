package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup

	ch := make(chan int, 5)
	quit := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 30; i++ {
			if i%5 == 0 {
				time.Sleep(time.Second * 2)
			}
			ch <- i
		}
		//close(ch)
		quit <- struct{}{}

		fmt.Println("goroutine 1 quit")
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case n := <-ch:
				fmt.Println(n)
				//forChanWithBreak(ch)
				forChanWithLen(ch)
				fmt.Println("finish one loop")
			case <-quit:
				return
			}
		}
	}()
	wg.Wait()
}

func rangeChan(ch chan int) {
	// range chan 只会在close(ch)后退出
	for n := range ch {
		fmt.Printf("--%d\n", n)
	}
}

func forChanWithBreak(ch chan int) {
	for {
		// 一直读取，知道ch关闭，ok为false
		n, ok := <-ch
		if !ok {
			break
		}
		fmt.Printf("**%d\n", n)
	}
}

func forChanWithLen(ch chan int) {
	// 尽可能多读
	for i := 0; i < len(ch); i++ {
		fmt.Printf("##%d\n", <-ch)
	}
}
