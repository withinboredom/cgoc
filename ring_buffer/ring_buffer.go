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

// WriteFull writes a message to the ring buffer ensuring the entire message is written
func (rb *RingBufferWrapper) WriteFull(data []byte) {
	C.ring_buffer_write_full(rb.ringBuffer, unsafe.Pointer(&data[0]), C.size_t(len(data)))
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
	var bytesRead int

	C.pthread_mutex_lock(&br.ringBuffer.mutex)
	defer C.pthread_mutex_unlock(&br.ringBuffer.mutex)

	deadlockCounter := 0

	for {
		if br.ringBuffer.read_index == br.ringBuffer.write_index {
			C.pthread_mutex_unlock(&br.ringBuffer.mutex)
			// yield the thread
			time.Sleep(1 * time.Microsecond)
			C.pthread_mutex_lock(&br.ringBuffer.mutex)
			if deadlockCounter < 1000 {
				deadlockCounter += 1
				continue
			}
			return nil, fmt.Errorf("no data available")
		}

		deadlockCounter = 0

		slot := C.ring_buffer_get_read_slot(br.ringBuffer)
		totalLength = int(slot.total_length)

		// Determine how much data to copy from the current slot
		remainingLength := totalLength - bytesRead
		copyLength := remainingLength
		if copyLength > C.MAX_DATA_LENGTH {
			copyLength = C.MAX_DATA_LENGTH
		}

		// Copy only the relevant portion of data from the current slot
		data := C.GoBytes(unsafe.Pointer(&slot.data[0]), C.int(copyLength))
		buffer.Write(data)

		fmt.Printf("\rRead %d bytes of %d\t\t", len(data), remainingLength)

		bytesRead += copyLength

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
	bufferReader := NewBufferReader(ringBuffer)

	for {
		wait_for_signal(ringBuffer) // Wait for the signal from C

		data, err := bufferReader.Read()
		if err != nil {
			fmt.Println("Read error:", err)
			continue
		}

		if output {
			fmt.Printf("\nReceived output: %s\n", string(data))
		} else {
			fmt.Printf("\nReceived data: %d\n", len(data))
			done()
		}
	}
}

func RunTest() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Create multiple ring buffers
	ringBuffer1 := NewRingBufferWrapper(10)
	defer ringBuffer1.Destroy()

	ringBuffer2 := NewRingBufferWrapper(5024)
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
	data := make([]byte, 1024*1024*1024)
	for i := 0; i < len(data); i++ {
		data[i] = byte(i % 256)
	}
	fmt.Printf("Writing large string: %d\n", len(data))

	// start a benchmarking timer
	start := time.Now()
	ringBuffer2.WriteFull(data)
	fmt.Println("Completed, waiting for read...")
	<-ctx.Done()
	end := time.Now()
	elapsed := end.Sub(start)
	fmt.Printf("\nTime taken for 2gb: %s\n", elapsed)

	select {}
}
