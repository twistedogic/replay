package replay

import (
	"errors"
	"io"
	"sync"
)

var ErrBufferFull = errors.New("buffer is full")

type ResetFunc func() (io.Writer, error)

type Writer struct {
	sync.Mutex
	buf []byte
	n   int
	w   io.Writer
	r   ResetFunc
}

func New(reset ResetFunc, size int) (*Writer, error) {
	buf := make([]byte, size)
	w := &Writer{
		buf: buf,
		r:   reset,
	}
	err := w.reset()
	go func() {
		for {
			w.writeAndRetry()
		}
	}()
	return w, err
}

func (w *Writer) reset() error {
	w.Lock()
	defer w.Unlock()
	nw, err := w.r()
	w.w = nw
	return err
}

func (w *Writer) write() error {
	w.Lock()
	defer w.Unlock()
	if w.n == 0 {
		return nil
	}
	n, err := w.w.Write(w.buf[:w.n])
	if n < w.n && err == nil {
		return io.ErrShortWrite
	}
	if n > 0 && n < w.n {
		copy(w.buf[0:w.n-n], w.buf[n:w.n])
	}
	w.n -= n
	return err
}

func (w *Writer) writeAndRetry() error {
	if w.w == nil {
		return w.reset()
	}
	if err := w.write(); err != nil {
		return w.reset()
	}
	return nil
}

func (w *Writer) Write(b []byte) (int, error) {
	w.Lock()
	defer w.Unlock()
	if len(b)+w.n > len(w.buf) {
		return 0, ErrBufferFull
	}
	n := copy(w.buf[w.n:], b)
	w.n += n
	return n, nil
}
