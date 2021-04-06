package tunnelify

import (
	"bytes"
	"io"
)

//TODO implement the close method

// Read wrapper is a wrapper around an io.ReadWriter
// Writing to this, writes to the underlying ReadWriter ,
// while reading from it, reads from the buffer first, before reading the underlying io.ReadWriter
// once the buffer is empty it can't be read from again. See io.MultiReader
type ConnectionReader struct {
	multiReader io.Reader
	reader      io.Reader
	writer      io.Writer
}

func NewConnectionReader(prepend []byte, reader io.Reader, writer io.Writer) *ConnectionReader {
	cw := ConnectionReader{
		reader: reader,
		writer: writer,
	}
	cw.multiReader = io.MultiReader(bytes.NewBuffer(prepend), cw.reader)
	return &cw
}

func (cr *ConnectionReader) Write(p []byte) (int, error) {
	return cr.writer.Write(p)
}

func (cr *ConnectionReader) Read(p []byte) (int, error) {
	return cr.multiReader.Read(p)
}

func (cw *ConnectionReader) Close() {
}
