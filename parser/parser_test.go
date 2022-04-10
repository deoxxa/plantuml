package parser

import (
  "bytes"
  "io/ioutil"
  "testing"

  "github.com/stretchr/testify/assert"
)

var testFileCache map[string][]byte

func readTestFile(name string) []byte {
  if testFileCache == nil {
    testFileCache = make(map[string][]byte)
  }

  if d, ok := testFileCache[name]; ok {
    return d
  }

  if d, err := ioutil.ReadFile("testdata/" + name); err == nil {
    testFileCache[name] = d
    return d
  }

  return nil
}

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

func TestParser(t *testing.T) {
  a := assert.New(t)

  doc, err := parseDocument(&scanner{d: readTestFile("simple-code-1-input.uml")})
  a.NoError(err)
  a.NotNil(doc)

  a.Equal("Begin", doc.FindInitialState().Name)
  a.Equal("Value1", doc.GetSkinParam("Param1"))
  a.Equal("Value2", doc.GetSkinParam("Param2"))
  a.Equal("", doc.GetSkinParam("Param3"))
}

func BenchmarkParser(b *testing.B) {
  for i := 0; i < b.N; i++ {
    parseDocument(&scanner{d: readTestFile("simple-code-1-input.uml")})
  }
}

func TestFormatter(t *testing.T) {
  for _, e := range []struct {
    name          string
    input, output []byte
  }{
    // {"simple", simpleCode, simpleCodeFormatted},
    // {"complex", complexCode, complexCodeFormatted},
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

func BenchmarkFormatter(b *testing.B) {
  doc, _ := parseDocument(&scanner{d: readTestFile("simple-code-1-input.uml")})

  for i := 0; i < b.N; i++ {
    FormatDocument(*doc, ioutil.Discard)
  }
}
