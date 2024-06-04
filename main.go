package main

// #include <stdio.h>
// #include <unistd.h>
// #include "ring_buffer/ring_buffer.h"
// void example(int mbs, RingBuffer *buffer);
// #include "example.c"
import "C"

import (
	"fmt"
	"github.com/withinboredom/cgoc/ring_buffer"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
	"unsafe"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: cgoc <mbs>")
		return
	}

	num, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Error:", err.Error())
		return
	}

	fmt.Println("Handing off to C to generate and send data!")
	runtime.LockOSThread()

	buffer := ring_buffer.NewBuffer(10)
	wg := sync.WaitGroup{}
	start := time.Now()

	wg.Add(1)
	go func() {
		data := <-buffer.Read()
		end := time.Now()
		fmt.Printf("Read %d bytes...\nPreview: %s...\n", len(data), string(data[:64]))
		fmt.Printf("Avg speed: %0.2f gb/s\n", float64(num)/1024/end.Sub(start).Seconds())
		fmt.Printf("Time elapsed: %s\n", end.Sub(start))
		wg.Done()
	}()

	C.example(C.int(num), (*C.RingBuffer)(unsafe.Pointer(buffer.RingBuffer)))

	wg.Wait()
}
