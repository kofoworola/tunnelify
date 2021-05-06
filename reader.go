package main

import (
	"bytes"
	"io"
	"net"
)

// Read wrapper is a wrapper around an io.ReadWriter
// Writing to this, writes to the underlying ReadWriter ,
// while reading from it, reads from the buffer first, before reading the underlying io.ReadWriter
// once the buffer is empty it can't be read from again. See io.MultiReader
type ConnectionReader struct {
	net.Conn

	multiReader io.Reader
}

func NewConnectionReader(rw net.Conn, prepend ...[]byte) *ConnectionReader {
	cw := ConnectionReader{
		Conn: rw,
	}
	readers := make([]io.Reader, len(prepend)+1)
	for i, item := range prepend {
		readers[i] = bytes.NewBuffer(item)
	}
	readers[len(readers)-1] = rw
	cw.multiReader = io.MultiReader(readers...)
	return &cw
}

func (cr *ConnectionReader) Read(p []byte) (int, error) {
	return cr.multiReader.Read(p)
}
