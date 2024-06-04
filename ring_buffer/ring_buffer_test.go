package ring_buffer

import (
	"bytes"
	"context"
	"testing"
	"time"
)

func TestRingBufferWriteAndRead(t *testing.T) {
	rb := newBufferWrapper(10)
	defer rb.destroy()

	data := []byte("test data")
	rb.writeFull(data)

	bufferReader := newBufferReader(rb.RingBuffer)
	readData, err := bufferReader.read()
	if err != nil {
		t.Fatalf("Failed to read data: %v", err)
	}

	if !bytes.Equal(data, readData) {
		t.Errorf("Data mismatch. Expected: %s, Got: %s", string(data), string(readData))
	}
}

func TestRingBufferLargeWriteAndRead(t *testing.T) {
	rb := newBufferWrapper(100)
	defer rb.destroy()

	// Create large data of 1MB
	data := make([]byte, 1024*1024)
	for i := 0; i < len(data); i++ {
		data[i] = byte(i % 256)
	}

	rb.writeFull(data)

	bufferReader := newBufferReader(rb.RingBuffer)
	readData, err := bufferReader.read()
	if err != nil {
		t.Fatalf("Failed to read data: %v", err)
	}

	if !bytes.Equal(data, readData) {
		t.Errorf("Data mismatch. Lengths: Expected %d, Got %d", len(data), len(readData))
	}
}

func TestConcurrentWriteAndRead(t *testing.T) {
	rb := newBufferWrapper(10)
	defer rb.destroy()

	data := []byte("concurrent data")

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		rb.writeFull(data)
	}()

	bufferReader := newBufferReader(rb.RingBuffer)
	readDataCh := make(chan []byte, 1)

	go func() {
		readData, err := bufferReader.read()
		if err != nil {
			t.Fatalf("Failed to read data: %v", err)
		}
		readDataCh <- readData
	}()

	select {
	case readData := <-readDataCh:
		if !bytes.Equal(data, readData) {
			t.Errorf("Data mismatch. Expected: %s, Got: %s", string(data), string(readData))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for read data")
	}
}

func TestRingBufferSynchronization(t *testing.T) {
	rb := newBufferWrapper(10)
	defer rb.destroy()

	data := []byte("sync test")

	// Start writer
	go func() {
		time.Sleep(500 * time.Millisecond)
		rb.writeFull(data)
	}()

	// Start reader
	bufferReader := newBufferReader(rb.RingBuffer)
	readData, err := bufferReader.read()
	if err != nil {
		t.Fatalf("Failed to read data: %v", err)
	}

	if !bytes.Equal(data, readData) {
		t.Errorf("Data mismatch. Expected: %s, Got: %s", string(data), string(readData))
	}
}
