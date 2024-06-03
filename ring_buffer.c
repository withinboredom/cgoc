#include "ring_buffer.h"

void ring_buffer_init(RingBuffer *ring_buffer, RingBufferSlot *slots, size_t size) {
    ring_buffer->slots = slots;
    ring_buffer->write_index = 0;
    ring_buffer->read_index = 0;
    ring_buffer->size = size;
    pthread_mutex_init(&ring_buffer->mutex, NULL);
    pthread_cond_init(&ring_buffer->cond, NULL);
}

void ring_buffer_destroy(RingBuffer *ring_buffer) {
    pthread_mutex_destroy(&ring_buffer->mutex);
    pthread_cond_destroy(&ring_buffer->cond);
}

int ring_buffer_write(RingBuffer *ring_buffer, const void *data, size_t data_length) {
    pthread_mutex_lock(&ring_buffer->mutex);

    // todo: prevent overflowing buffer
    size_t remaining = data_length;
    size_t offset = 0;

    while (remaining > 0) {
        RingBufferSlot *slot = &ring_buffer->slots[ring_buffer->write_index];
        size_t chunk_size = remaining > MAX_DATA_LENGTH ? MAX_DATA_LENGTH : remaining;

        slot->total_length = data_length;
        slot->fragment_offset = offset;
        memcpy(slot->data, (char *)data + offset, chunk_size);

        remaining -= chunk_size;
        offset += chunk_size;
        ring_buffer->write_index = (ring_buffer->write_index + 1) % ring_buffer->size;

        pthread_cond_signal(&ring_buffer->cond); // Signal that new data is available
    }

    pthread_mutex_unlock(&ring_buffer->mutex);

    return data_length;
}

RingBufferSlot* get_read_slot(RingBuffer *ring_buffer) {
    return &ring_buffer->slots[ring_buffer->read_index];
}

void advance_read_index(RingBuffer *ring_buffer) {
    ring_buffer->read_index = (ring_buffer->read_index + 1) % ring_buffer->size;
}

void wait_for_signal(RingBuffer *ring_buffer) {
    pthread_mutex_lock(&ring_buffer->mutex);
    while (ring_buffer->write_index == ring_buffer->read_index) {
        pthread_cond_wait(&ring_buffer->cond, &ring_buffer->mutex);
    }
    pthread_mutex_unlock(&ring_buffer->mutex);
}
