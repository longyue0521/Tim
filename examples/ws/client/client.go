package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func test(i int, wg *sync.WaitGroup) {
	conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), "ws://127.0.0.1:2000/")
	defer wg.Done()
	if err != nil {
		fmt.Printf("%d can not connect: %v\n", i, err)
	} else {
		fmt.Printf("%d connected\n", i)
		msg := []byte("OK+OK")
		err = wsutil.WriteClientMessage(conn, ws.OpText, msg)
		if err != nil {
			fmt.Printf("%d can not send: %v\n", i, err)
			return
		} else {
			fmt.Printf("%d send: %s, type: %v\n", i, msg, ws.OpText)
		}

		msg, op, err := wsutil.ReadServerData(conn)
		if err != nil {
			fmt.Printf("%d can not receive: %v\n", i, err)
			return
		} else {
			fmt.Printf("%d receive: %s，type: %v\n", i, msg, op)
		}

		time.Sleep(time.Duration(3) * time.Second)

		err = conn.Close()
		if err != nil {
			fmt.Printf("%d can not close: %v\n", i, err)
		} else {
			fmt.Printf("%d closed\n", i)
		}
	}
}

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 1; i++ {
		wg.Add(1)
		go test(i, &wg)
	}
	wg.Wait()
}
