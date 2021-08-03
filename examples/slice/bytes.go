package main

import "fmt"

func main() {
	var b []byte
	bb := []byte(nil)
	bbb := []byte{}
	bbbb := make([]byte, 0)

	fmt.Println(b, bb, bbb, bbbb)
	fmt.Println(b == nil, bb == nil, bbb == nil, bbbb == nil)
	fmt.Println(len(b), len(bb), len(bbb), len(bbbb))
	// fmt.Println(b == []byte(nil))
	// invalid operation: b == ([]byte)(nil) (slice can only be compared to nil)
	//
}
