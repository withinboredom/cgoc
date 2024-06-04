package ring_buffer

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"
)

func TestWriting(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Create multiple ring buffers
	ringBuffer1 := NewRingBufferWrapper(10)
	defer ringBuffer1.Destroy()

	ringBuffer2 := NewRingBufferWrapper(10)
	defer ringBuffer2.Destroy()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// Example of concurrent writers
	go streamData(ringBuffer1.ringBuffer, true, cancel)
	go streamData(ringBuffer2.ringBuffer, false, cancel)

	/*for i := 0; i < 10; i++ {
		output := []byte("example output " + strconv.Itoa(i))
		ringBuffer1.Write(output)
	}*/

	// create a 2gb string
	fmt.Println("Generating large string!")
	data := make([]byte, 1*1024*1024*1024)
	for i := 0; i < len(data); i++ {
		data[i] = byte(i % 256)
	}
	fmt.Println("Writing large string!")

	// start a benchmarking timer
	start := time.Now()
	ringBuffer2.WriteFull(data)
	fmt.Println("Completed, waiting for read...")
	<-ctx.Done()
	end := time.Now()
	elapsed := end.Sub(start)
	fmt.Printf("Time taken for 2gb: %s\n", elapsed)

	select {}
}
