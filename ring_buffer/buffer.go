package ring_buffer

import (
	"sync"
)

type Buffer struct {
	RingBufferWrapper
	readChan chan []byte
	stopChan chan struct{}
	wg       sync.WaitGroup
}

func NewBuffer(size int) *Buffer {
	rb := newBufferWrapper(size)
	b := &Buffer{
		RingBufferWrapper: *rb,
		readChan:          make(chan []byte),
		stopChan:          make(chan struct{}),
		wg:                sync.WaitGroup{},
	}

	b.wg.Add(1)
	go b.startReadRoutine()

	return b
}

func (b *Buffer) Destroy() {
	close(b.stopChan)
	b.wg.Wait()

	b.RingBufferWrapper.Destroy()
}

func (b *Buffer) Write(data []byte) int {
	return b.write(data)
}

func (b *Buffer) Read() chan []byte {
	return b.readChan
}

func (b *Buffer) startReadRoutine() {
	defer b.wg.Done()

	reader := newBufferReader(b.RingBuffer)

	for {
		select {
		case <-b.stopChan:
			return
		default:
			wait_for_signal(b.RingBuffer)
			data, err := reader.read()
			if err != nil {
				panic(err)
			}

			b.readChan <- data
		}
	}
}
