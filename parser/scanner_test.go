package parser

import (
  "testing"

  "github.com/stretchr/testify/assert"
)

func TestScanner(t *testing.T) {
  a := assert.New(t)

  const code = "12 \t \t \t 34"

  s := &scanner{d: []byte(code)}

  a.Equal(0, s.pos())
  a.Equal(byte('1'), s.peek())
  a.Equal(0, s.pos())
  a.Equal(byte('1'), s.byte())
  a.Equal(1, s.pos())
  a.Equal(byte('2'), s.byte())
  a.Equal(2, s.pos())
  s.ws()
  a.Equal(9, s.pos())
  a.Equal(byte('3'), s.byte())
}

func BenchmarkScanner(b *testing.B) {
  const code = "12 \t \t \t 34"

  for i := 0; i < b.N; i++ {
    s := &scanner{d: []byte(code)}
    s.byte()
    s.byte()
    s.ws()
    s.byte()
    s.byte()
  }
}
