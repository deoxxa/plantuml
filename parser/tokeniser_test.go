package parser

import (
  "testing"

  "github.com/stretchr/testify/assert"
)

func TestTokeniser(t *testing.T) {
  a := assert.New(t)

  var i int
  s := &scanner{d: readTestFile("simple-code-1-input.uml")}
  for getToken(s, &options{parseTrailing: true}) != nil {
    i++
  }

  a.Equal(64, i)
}

func BenchmarkTokeniser(b *testing.B) {
  for i := 0; i < b.N; i++ {
    s := &scanner{d: readTestFile("simple-code-1-input.uml")}
    for getToken(s, nil) != nil {
    }
  }
}
