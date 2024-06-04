#include <stdio.h>

#define MB 1024 * 1024

void example(int mbs, RingBuffer *buffer) {
  char *bytes = malloc(mbs * MB);  // Allocating the memory

  // Don't forget to check for null pointer in case the memory allocation fails.
  if (bytes == NULL) {
    // Memory allocation failed.
    printf("Memory allocation failed!\n");
    return;
  }

  memset(bytes, 'A', mbs * MB);

  printf("Sending bytes...\n");
  ring_buffer_write_full(buffer, bytes, mbs * MB);
  printf("Done sending bytes!\n");

  sleep(1);

  free(bytes);
}