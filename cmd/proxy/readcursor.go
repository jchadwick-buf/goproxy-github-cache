package main

import (
	"io"
)

// readCursor is a wrapper around a io.ReaderAt that provides an io.Reader.
type readCursor struct {
	readerAtCloser
	cursor int64
}

// Constructs a new readCursor at the start of the reader.
func newReadCursor(readerAtCloser readerAtCloser) io.ReadCloser {
	return &readCursor{readerAtCloser: readerAtCloser}
}

// Read implements io.ReadSeekCloser.
func (reader *readCursor) Read(p []byte) (n int, err error) {
	n, err = reader.ReadAt(p, reader.cursor)
	reader.cursor += int64(n)
	return n, err
}

type readerAtCloser interface {
	io.ReaderAt
	io.Closer
}
