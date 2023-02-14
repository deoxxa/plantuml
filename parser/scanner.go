package parser

import (
	"bytes"
	"fmt"
)

type scanner struct {
	d []byte
	p int
	h []int
	a [][]SourceRange
}

func (s *scanner) pos() int    { return s.p }
func (s *scanner) savePos()    { s.h = append(s.h, s.p) }
func (s *scanner) restorePos() { s.p = s.h[len(s.h)-1]; s.discardPos() }
func (s *scanner) discardPos() { s.h = s.h[:len(s.h)-1] }

func (s *scanner) moveTo(tk *token)        { s.p = tk.pos[0] }
func (s *scanner) peek() byte              { return s.d[s.p] }
func (s *scanner) move(n int)              { s.p += n }
func (s *scanner) byte() byte              { b := s.d[s.p]; s.move(1); return b }
func (s *scanner) read(n int) []byte       { d := s.d[s.p : s.p+n]; s.move(n); return d }
func (s *scanner) readString(n int) string { return string(s.read(n)) }

func (s *scanner) eof() bool { return s.p >= len(s.d) }

func (s *scanner) ws() {
	for !s.eof() {
		switch s.peek() {
		case ' ', '\t':
			s.move(1)
		default:
			return
		}
	}
}

func (s *scanner) wsnl() {
	for !s.eof() {
		switch s.peek() {
		case ' ', '\t', '\r', '\n':
			s.move(1)
		default:
			return
		}
	}
}

func (s *scanner) nl() {
	for !s.eof() {
		switch s.peek() {
		case '\r', '\n':
			s.move(1)
		default:
			return
		}
	}
}

func GetLineAndColumnForOffset(d []byte, p int) [2]int {
	b := d[0:p]
	l := bytes.Count(b, []byte("\n"))
	i := bytes.LastIndex(b, []byte("\n"))
	return [2]int{l + 1, len(b) - i}
}

func (s *scanner) lc(p int) [2]int {
	return GetLineAndColumnForOffset(s.d, p)
}

func (s *scanner) tsp(p int) SourcePosition {
	l := s.lc(p)
	return SourcePosition{Offset: p, Line: l[0], Column: l[1]}
}

func (s *scanner) tsr(tk *token) SourceRange {
	return SourceRange{
		Start: s.tsp(tk.pos[0]),
		End:   s.tsp(tk.pos[1]),
	}
}

func (s *scanner) pushTrackedRange() {
	s.a = append(s.a, []SourceRange{})
}

func (s *scanner) popTrackedRange() SourceRange {
	a := s.a[len(s.a)-1]
	s.a = s.a[0 : len(s.a)-1]
	return MergeRanges(a)
}

func (s *scanner) trackTokenRange(tk *token) {
	r := s.tsr(tk)
	for i := range s.a {
		s.a[i] = append(s.a[i], r)
	}
}

func (s *scanner) err(err error) error {
	lc := s.lc(s.p)
	return fmt.Errorf("error at %d:%d: %w", lc[0], lc[1], err)
}

func (s *scanner) rerr(err error) error {
	e := s.err(err)
	s.restorePos()
	return e
}
