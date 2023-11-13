package main

import (
	"errors"
	"io"

	actionscache "github.com/tonistiigi/go-actions-cache"
)

type readerBlob struct {
	io.ReadSeeker
	size int64
}

// newReaderBlob constructs a new actionscache.Blob from an io.ReadSeeker.
func newReaderBlob(reader io.ReadSeeker) (actionscache.Blob, error) {
	// Determine size of io.ReadSeeker
	pos, err := reader.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	size, err := reader.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}
	_, err = reader.Seek(pos, io.SeekStart)
	if err != nil {
		return nil, err
	}

	return &readerBlob{
		ReadSeeker: reader,
		size:       size,
	}, nil
}

// ReadAt implements io.ReaderAt
func (b *readerBlob) ReadAt(p []byte, off int64) (n int, err error) {
	pos, err := b.ReadSeeker.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}
	defer func() {
		_, err2 := b.ReadSeeker.Seek(pos, io.SeekStart)
		err = errors.Join(err, err2)
	}()
	_, err = b.ReadSeeker.Seek(off, io.SeekStart)
	if err != nil {
		return 0, err
	}
	n, err = b.ReadSeeker.Read(p)
	return n, err
}

// Size implements actionscache.Blob
func (b *readerBlob) Size() int64 {
	return b.size
}

// Size implements actionscache.Blob
func (b *readerBlob) Close() error {
	return nil
}
