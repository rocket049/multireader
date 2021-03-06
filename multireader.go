// Author: Fu Huizhong<fuhuizn@163.com>
package multireader

import (
	"io"
	"sync/atomic"
)

//RandMultiReader This struct looks like io.MultiReader,but can Read input data from every reader in real time。
type RandMultiReader struct {
	channel chan []byte
	buf     []byte
	num     int32
}

//Read compatible with io.Reader
func (s *RandMultiReader) Read(p []byte) (n int, err error) {
	lbuf := len(s.buf)
	lp := len(p)
	if lbuf > 0 {
		if lbuf <= lp {
			n = copy(p, s.buf)
			s.buf = nil
		} else {
			n = copy(p, s.buf[:lp])
			s.buf = s.buf[lp:]
		}
	} else {
		var ok bool
		s.buf, ok = <-s.channel
		if !ok {
			return 0, io.EOF
		} else {
			lbuf = len(s.buf)
			if lbuf <= lp {
				n = copy(p, s.buf)
				s.buf = nil
			} else {
				n = copy(p, s.buf[:lp])
				s.buf = s.buf[lp:]
			}
		}
	}
	return
}

func (s *RandMultiReader) linkReader(r io.ReadCloser) error {
	defer r.Close()
	for {
		p := make([]byte, 512)
		n, err := r.Read(p)
		if err != nil {
			ret := atomic.AddInt32(&s.num, -1)
			if ret == 0 {
				close(s.channel)
			}
			return err
		}
		s.channel <- p[:n]
	}
}

//NewRandMultiReader Create a new RandMultiReader from sevel readers.
func NewRandMultiReader(readers ...io.ReadCloser) io.Reader {
	res := &RandMultiReader{channel: make(chan []byte, 10), buf: nil, num: int32(len(readers))}
	for _, v := range readers {
		go res.linkReader(v)
	}
	return res
}
