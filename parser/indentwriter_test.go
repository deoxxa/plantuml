package parser

import (
  "bytes"
  "io"
  "io/ioutil"
  "testing"

  "github.com/stretchr/testify/assert"
)

func TestIndentWriter(t *testing.T) {
  for _, e := range []struct {
    name          string
    input, output []byte
  }{
    {"1", readTestFile("indentwriter-1-input.txt"), readTestFile("indentwriter-1-output.txt")},
  } {
    t.Run(e.name, func(t *testing.T) {
      a := assert.New(t)

      buf := bytes.NewBuffer(nil)

      wr := NewIndentWriter(buf, []byte("  "))

      n, err := io.Copy(wr, bytes.NewReader(e.input))
      a.NoError(err)
      a.Equal(int64(len(e.input)), n)
      a.Equal(string(e.output), buf.String())
    })
  }
}

func BenchmarkIndentWriter(b *testing.B) {
  for _, e := range []struct {
    name          string
    input, output []byte
  }{
    {"1", readTestFile("indentwriter-1-input.txt"), readTestFile("indentwriter-1-output.txt")},
  } {
    b.Run(e.name, func(b *testing.B) {
      for i := 0; i < b.N; i++ {
        io.Copy(NewIndentWriter(ioutil.Discard, []byte("  ")), bytes.NewReader(e.input))
      }
    })
  }
}
