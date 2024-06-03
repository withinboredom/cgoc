#ifndef RING_BUFFER_H
#define RING_BUFFER_H

#include <stdlib.h>
#include <string.h>
#include <pthread.h>

#define CACHE_LINE_SIZE 64
#define MAX_DATA_LENGTH 64

typedef struct {
    size_t total_length;
    size_t fragment_offset;
    char data[MAX_DATA_LENGTH];
} __attribute__((aligned(CACHE_LINE_SIZE))) RingBufferSlot;

typedef struct {
    RingBufferSlot *slots;
    size_t write_index;
    size_t read_index;
    size_t size;
    pthread_mutex_t mutex;
    pthread_cond_t cond;
} RingBuffer;

void ring_buffer_init(RingBuffer *ring_buffer, RingBufferSlot *slots, size_t size);
void ring_buffer_destroy(RingBuffer *ring_buffer);
int ring_buffer_write(RingBuffer *ring_buffer, const void *data, size_t data_length);
RingBufferSlot* ring_buffer_get_read_slot(RingBuffer *ring_buffer);
void ring_buffer_advance_read_index(RingBuffer *ring_buffer);
void ring_buffer_wait_for_signal(RingBuffer *ring_buffer);

#endif
