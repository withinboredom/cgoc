package ring_buffer

/*
#include <stdlib.h>
#include "ring_buffer.h"
*/
import "C"
import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
	"unsafe"
)

// RingBufferWrapper wraps the C RingBuffer struct
type RingBufferWrapper struct {
	RingBuffer *C.RingBuffer
	slots      []C.RingBufferSlot
}

// newBufferWrapper creates a new RingBufferWrapper
func newBufferWrapper(size int) *RingBufferWrapper {
	slots := make([]C.RingBufferSlot, size)
	rb := &RingBufferWrapper{
		RingBuffer: (*C.RingBuffer)(C.malloc(C.sizeof_RingBuffer)),
		slots:      slots,
	}
	C.ring_buffer_init(rb.RingBuffer, (*C.RingBufferSlot)(unsafe.Pointer(&slots[0])), C.size_t(size))
	return rb
}

// destroy releases the resources used by the RingBufferWrapper
func (rb *RingBufferWrapper) Destroy() {
	C.ring_buffer_destroy(rb.RingBuffer)
	C.free(unsafe.Pointer(rb.RingBuffer))
}

// write writes a message to the ring buffer with fragmentation support
func (rb *RingBufferWrapper) write(data []byte) int {
	written := C.ring_buffer_write(rb.RingBuffer, unsafe.Pointer(&data[0]), C.size_t(len(data)))
	return int(written)
}

// writeFull writes a message to the ring buffer ensuring the entire message is written
func (rb *RingBufferWrapper) writeFull(data []byte) {
	C.ring_buffer_write_full(rb.RingBuffer, unsafe.Pointer(&data[0]), C.size_t(len(data)))
}

// bufferReader struct to encapsulate ring buffer reading
type bufferReader struct {
	mu         sync.Mutex
	ringBuffer *C.RingBuffer
}

// newBufferReader creates a new bufferReader
func newBufferReader(ringBuffer *C.RingBuffer) *bufferReader {
	return &bufferReader{
		ringBuffer: ringBuffer,
	}
}

const minWait = 1 * time.Microsecond
const maxWait = 1 * time.Millisecond

// read reads a message from the ring buffer, handling fragmentation
func (br *bufferReader) read() ([]byte, error) {
	br.mu.Lock()
	defer br.mu.Unlock()

	var buffer bytes.Buffer
	var totalLength int
	var bytesRead int

	//time.Sleep(1 * time.Millisecond)
	C.pthread_mutex_lock(&br.ringBuffer.mutex)
	defer C.pthread_mutex_unlock(&br.ringBuffer.mutex)

	for {
		//time.Sleep(1 * time.Millisecond)
		if br.ringBuffer.read_index == br.ringBuffer.write_index {
			delay := minWait
			C.pthread_mutex_unlock(&br.ringBuffer.mutex)
			for br.ringBuffer.read_index == br.ringBuffer.write_index && delay <= maxWait {
				time.Sleep(delay)
				delay *= 2
			}
			if br.ringBuffer.read_index == br.ringBuffer.write_index {
				time.Sleep(maxWait)
			}
			C.pthread_mutex_lock(&br.ringBuffer.mutex)
			continue
		}

		slot := C.ring_buffer_get_read_slot(br.ringBuffer)
		totalLength = int(slot.total_length)

		remainingLength := totalLength - bytesRead
		copyLength := remainingLength
		if copyLength > C.MAX_DATA_LENGTH {
			copyLength = C.MAX_DATA_LENGTH
		}

		// create a slice
		slice := unsafe.Slice((*byte)(unsafe.Pointer(slot.data)), copyLength)
		data := make([]byte, len(slice))
		copy(data, slice)

		// Copy only the relevant portion of data from the current slot
		//data := C.GoBytes(unsafe.Pointer(slot.data), C.int(copyLength))
		buffer.Write(data)

		//fmt.Printf("\rRead %d bytes of %d\t\t", len(data), remainingLength)

		bytesRead += len(data)

		C.ring_buffer_advance_read_index(br.ringBuffer)

		if bytesRead >= totalLength {
			return buffer.Bytes(), nil
		}
	}
}

func wait_for_signal(ringBuffer *C.RingBuffer) {
	C.ring_buffer_wait_for_signal(ringBuffer)
}

func streamData(ringBuffer *C.RingBuffer, output bool, done context.CancelFunc) {
	bufferReader := newBufferReader(ringBuffer)

	for {
		wait_for_signal(ringBuffer) // Wait for the signal from C

		data, err := bufferReader.read()
		if err != nil {
			fmt.Println("read error:", err)
			continue
		}

		if output {
			fmt.Printf("\nReceived output: %s\n", string(data))
		} else {
			fmt.Printf("\nReceived data: %d\n", len(data))
			done()
			return
		}
	}
}

func RunTest() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ringBuffer2 := newBufferWrapper(100)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	//go streamData(ringBuffer1.ringBuffer, true, cancel)
	go streamData(ringBuffer2.RingBuffer, false, cancel)

	fmt.Println("Generating large string!")
	data := make([]byte, 2*1024*1024*1024)
	for i := 0; i < len(data); i++ {
		data[i] = byte(i % 256)
	}
	fmt.Println("Writing large string!")

	start := time.Now()
	ringBuffer2.writeFull(data)
	fmt.Println("Completed, waiting for read...")
	<-ctx.Done()
	end := time.Now()
	elapsed := end.Sub(start)
	fmt.Printf("Time taken for 1GB: %s | %0.2f gb/s\n", elapsed, 1/elapsed.Seconds())
}
