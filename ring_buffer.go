package cgoc

/*
#include <stdlib.h>
#include "ring_buffer.h"
*/
import "C"
import (
	"bytes"
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

// RingBufferWrapper wraps the C RingBuffer struct
type RingBufferWrapper struct {
	ringBuffer *C.RingBuffer
	slots      []C.RingBufferSlot
}

// NewRingBufferWrapper creates a new RingBufferWrapper
func NewRingBufferWrapper(size int) *RingBufferWrapper {
	slots := make([]C.RingBufferSlot, size)
	rb := &RingBufferWrapper{
		ringBuffer: (*C.RingBuffer)(C.malloc(C.sizeof_RingBuffer)),
		slots:      slots,
	}
	C.ring_buffer_init(rb.ringBuffer, (*C.RingBufferSlot)(unsafe.Pointer(&slots[0])), C.size_t(size))
	return rb
}

// Destroy releases the resources used by the RingBufferWrapper
func (rb *RingBufferWrapper) Destroy() {
	C.ring_buffer_destroy(rb.ringBuffer)
	C.free(unsafe.Pointer(rb.ringBuffer))
}

// Write writes a message to the ring buffer with fragmentation support
func (rb *RingBufferWrapper) Write(data []byte) int {
	written := C.ring_buffer_write(rb.ringBuffer, unsafe.Pointer(&data[0]), C.size_t(len(data)))
	return int(written)
}

// BufferReader struct to encapsulate ring buffer reading
type BufferReader struct {
	mu         sync.Mutex
	ringBuffer *C.RingBuffer
}

// NewBufferReader creates a new BufferReader
func NewBufferReader(ringBuffer *C.RingBuffer) *BufferReader {
	return &BufferReader{
		ringBuffer: ringBuffer,
	}
}

// Read reads a message from the ring buffer, handling fragmentation
func (br *BufferReader) Read() ([]byte, error) {
	br.mu.Lock()
	defer br.mu.Unlock()

	var buffer bytes.Buffer
	var totalLength int

	C.pthread_mutex_lock(&br.ringBuffer.mutex)
	defer C.pthread_mutex_unlock(&br.ringBuffer.mutex)

	for {
		if br.ringBuffer.read_index == br.ringBuffer.write_index {
			return nil, fmt.Errorf("no data available")
		}

		slot := C.ring_buffer_get_read_slot(br.ringBuffer)
		totalLength = int(slot.total_length)
		data := C.GoBytes(unsafe.Pointer(&slot.data[0]), C.int(slot.fragment_offset+slot.total_length))

		buffer.Write(data)

		C.ring_buffer_advance_read_index(br.ringBuffer)

		if buffer.Len() >= totalLength {
			return buffer.Bytes(), nil
		}
	}
}

func wait_for_signal(ringBuffer *C.RingBuffer) {
	C.ring_buffer_wait_for_signal(ringBuffer)
}

func streamData(ringBuffer *C.RingBuffer) {
	bufferReader := NewBufferReader(ringBuffer)

	for {
		wait_for_signal(ringBuffer) // Wait for the signal from C

		data, err := bufferReader.Read()
		if err != nil {
			fmt.Println("Read error:", err)
			continue
		}

		fmt.Printf("Received output: %s\n", string(data))
	}
}

func main() {
	// Ensure this goroutine stays on the same OS thread
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Create multiple ring buffers
	ringBuffer1 := NewRingBufferWrapper(10)
	defer ringBuffer1.Destroy()

	ringBuffer2 := NewRingBufferWrapper(10)
	defer ringBuffer2.Destroy()

	// Example of concurrent writers
	go streamData(ringBuffer1.ringBuffer)
	go streamData(ringBuffer2.ringBuffer)

	output := []byte("example output")

	ringBuffer1.Write(output)

	// Keep the main function running
	select {}
}
