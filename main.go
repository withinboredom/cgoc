package main

import (
	"cgoc/buffer"
	"fmt"
	"log"
	"runtime"
)

func main() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	rb := buffer.NewRingBuffer(10)
	defer rb.Destroy()

	go func() {
		for i := 0; i < 10; i++ {
			data := []byte(fmt.Sprintf("data %d", i))
			if err := rb.Write(data); err != nil {
				log.Fatalf("Failed to write data: %v", err)
			}
		}
	}()

	for data := range rb.Read() {
		fmt.Printf("Read data: %s\n", string(data))
	}
}
