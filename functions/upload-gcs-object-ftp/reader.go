package ftp

import (
	"context"
	"io"
)

type contextAwareReader struct {
	io.Reader
	ctx context.Context
}

func (r contextAwareReader) Read(p []byte) (n int, err error) {
	if err = r.ctx.Err(); err != nil {
		return
	}
	if n, err = r.Reader.Read(p); err != nil {
		return
	}
	err = r.ctx.Err()
	return
}
