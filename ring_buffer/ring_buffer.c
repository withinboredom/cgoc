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

int is_slot_unread(RingBuffer *ring_buffer, size_t index) {
    return (index + 1) % ring_buffer->size != ring_buffer->read_index;
}

size_t ring_buffer_write(RingBuffer *ring_buffer, const void *data, size_t data_length) {
    pthread_mutex_lock(&ring_buffer->mutex);

    size_t remaining = data_length;
    size_t offset = 0;

    while (remaining > 0) {
        size_t next_write_index = (ring_buffer->write_index + 1) % ring_buffer->size;

        if (!is_slot_unread(ring_buffer, next_write_index)) {
            pthread_cond_wait(&ring_buffer->cond, &ring_buffer->mutex);
        }

        RingBufferSlot *slot = &ring_buffer->slots[ring_buffer->write_index];
        size_t chunk_size = remaining > MAX_DATA_LENGTH ? MAX_DATA_LENGTH : remaining;

        slot->total_length = data_length;
        slot->fragment_offset = offset;
        slot->data = (const char *)data + offset;

        remaining -= chunk_size;
        offset += chunk_size;
        ring_buffer->write_index = next_write_index;

        pthread_cond_signal(&ring_buffer->cond);
    }

    pthread_mutex_unlock(&ring_buffer->mutex);

    return data_length;
}

void ring_buffer_write_full(RingBuffer *ring_buffer, const void *data, size_t data_length) {
    size_t total_written = 0;

    while (total_written < data_length) {
        size_t written = ring_buffer_write(ring_buffer, (const char *)data + total_written, data_length - total_written);
        total_written += written;
    }
}

RingBufferSlot* ring_buffer_get_read_slot(RingBuffer *ring_buffer) {
    return &ring_buffer->slots[ring_buffer->read_index];
}

void ring_buffer_advance_read_index(RingBuffer *ring_buffer) {
    ring_buffer->read_index = (ring_buffer->read_index + 1) % ring_buffer->size;
    pthread_cond_signal(&ring_buffer->cond);
}

void ring_buffer_wait_for_signal(RingBuffer *ring_buffer) {
    pthread_mutex_lock(&ring_buffer->mutex);
    while (ring_buffer->write_index == ring_buffer->read_index) {
        pthread_cond_wait(&ring_buffer->cond, &ring_buffer->mutex);
    }
    pthread_mutex_unlock(&ring_buffer->mutex);
}
