package downloads

import (
	"io"
)

type SectionWriter struct {
	w     io.WriterAt
	off   int64
	n     int64
	limit int64
}

// NewSectionWriter returns a SectionWriter that writes to w starting at offset off and stops with EOF after n bytes.
func NewSectionWriter(w io.WriterAt, off int64, n int64) *SectionWriter {
	return &SectionWriter{w, off, n, off + n}
}

func (s *SectionWriter) Write(p []byte) (n int, err error) {
	if s.off >= s.limit {
		return 0, io.EOF
	}
	if max := s.limit - s.off; int64(len(p)) > max {
		p = p[0:max]
	}
	n, err = s.w.WriteAt(p, s.off)
	s.off += int64(n)
	return
}
