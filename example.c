#include <stdio.h>
#include <unistd.h>
#include "example.h"

#define MB 1024 * 1024

void example(int mbs, RingBuffer *buffer) {
  char *bytes = malloc(mbs * MB);  // Allocating the memory

  printf("Using %d chunks\n", MAX_DATA_LENGTH);

  // Don't forget to check for null pointer in case the memory allocation fails.
  if (bytes == NULL) {
    // Memory allocation failed.
    printf("Memory allocation failed!\n");
    return;
  }

  memset(bytes, 'A', mbs * MB);
  for (int i = 0; i < mbs * MB; ++i) {
     bytes[i] = 'A' + (i % 26);
  }

  printf("Sending bytes...\n");
  ring_buffer_write_full(buffer, bytes, mbs * MB);
  printf("Done sending bytes!\n");

  free(bytes);
}