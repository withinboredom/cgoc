# CgoC (see-go-see)

This library provides a mechanism to send data from C to Go in the context of being inside a cgo application in order to
prevent issues with Go -> C -> Go. It works by using a circular buffer and pthread constructs (such as signals and
mutexes) to synchronize data between C -> Go. Essentially, this is a one-way channel from C to Go.

## Warning:

This is **experimental** at the moment and under heavy changes; do not use in production (yet)!

## Usage

Consider yourself warned! If you want to use this anyway, copy the [ring_buffer.h](./ring_buffer/ring_buffer.h) file
into your cgo project and import the module. The `main.go` file in this repository shows a way in which this can be
used.