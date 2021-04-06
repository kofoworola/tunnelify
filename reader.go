package tunnelify

import (
	"io"
	"sync"
)

//TODO check usefulness of this custom byteReader

// byteWrapper is a custom implementation of bytes.Reader
// that allows appending to the underlying data
type byteReader struct {
	s        []byte
	i        int64 // current reading index
	prevRune int   // index of previous rune; or < 0
}

func (bw *byteReader) Read(b []byte) (n int, err error) {
	val := bw.s
	if bw.i >= int64(len(val)) {
		return 0, io.EOF
	}
	bw.prevRune = -1
	n = copy(b, val[bw.i:])
	bw.i += int64(n)
	return
}

func (bw *byteReader) appendToData(b []byte) {
	bw.s = append(bw.s, b...)
}

// Read wrapper is a wrapper around an io.Reader
// It acts as an adapter to the reader and turns it to a io.ReadWriter
// Writing to this, writes to the buffer field,
// while reading from it, reads from the buffer first, before reading the underlying io.Reader
// once the buffer is empty it can't be read from again. See io.MultiReader
type ReadWrapper struct {
	wrapped     io.Reader
	buffer      byteReader
	lock        sync.RWMutex
	multiReader io.Reader
}

func NewReadWrapper(reader io.Reader) *ReadWrapper {
	rw := ReadWrapper{
		wrapped: reader,
		buffer: byteReader{
			s:        make([]byte, 0),
			i:        0,
			prevRune: -1,
		},
	}
	rw.multiReader = io.MultiReader(&rw.buffer, rw.wrapped)
	return &rw
}

func (r *ReadWrapper) Write(p []byte) (int, error) {
	r.lock.Lock()
	r.buffer.appendToData(p)
	r.lock.Unlock()
	return len(p), nil
}

func (r *ReadWrapper) Read(p []byte) (int, error) {
	r.lock.RLock()
	d, err := r.multiReader.Read(p)
	r.lock.RUnlock()
	return d, err
}
