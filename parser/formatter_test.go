package parser

import (
  "bytes"
  "io/ioutil"
  "testing"

  "github.com/stretchr/testify/assert"
)

func TestFormatDocument(t *testing.T) {
  for _, e := range []struct {
    name          string
    input, output []byte
  }{
    {"simple", readTestFile("simple-code-1-input.uml"), readTestFile("simple-code-1-formatted.uml")},
    {"complex", readTestFile("complex-code-1-input.uml"), readTestFile("complex-code-1-formatted.uml")},
    {"complex2", readTestFile("complex-code-2-input.uml"), readTestFile("complex-code-2-formatted.uml")},
  } {
    t.Run(e.name, func(t *testing.T) {
      a := assert.New(t)

      doc, err := parseDocument(&scanner{d: e.input})
      a.NoError(err)
      a.NotNil(doc)

      if doc != nil {
        buf := bytes.NewBuffer(nil)
        a.NoError(FormatDocument(*doc, buf))
        if e.output != nil {
          a.Equal(string(e.output), buf.String())
        }
      }
    })
  }
}

func BenchmarkFormatDocument(b *testing.B) {
  doc, _ := parseDocument(&scanner{d: readTestFile("simple-code-1-input.uml")})

  for i := 0; i < b.N; i++ {
    FormatDocument(*doc, ioutil.Discard)
  }
}
