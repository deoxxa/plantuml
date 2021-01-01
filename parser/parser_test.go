package parser

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

// this is a mess on purpose
const simpleCode = `
  @startuml
  skinparam     Param1 Value1
  skinparam Param2   Value2
  state "begin" as    Begin <<sdlreceive>> {
    state "Entry Condition 1" as Begin_E1 : FieldA == 0
    ---
    state   "Exit Condition 1" as     Begin_X1 : FieldA != 0
  }
  state "state-b" as StateB {
      state  "Exit Condition 1" as StateB_X1 : is(FieldB, 'value-a', 'value-v', 'value-c') AND !empty(FieldC)
               state  "Exit Condition 2" as StateB_X2 : is(FieldB, 'value-d') AND FieldD > 0
  }
  [*] --> Begin
  Begin    --> StateB : FieldE == 0
  @enduml
`

const simpleCodeFormatted = `@startuml
  skinparam Param1 Value1
  skinparam Param2 Value2

  state "begin" as Begin {
    state "Entry Condition 1" as Begin_E1 : FieldA == 0
    ---
    state "Exit Condition 1" as Begin_X1 : FieldA != 0
  }
  state "state-b" as StateB {
    state "Exit Condition 1" as StateB_X1 : is(FieldB, 'value-a', 'value-v', 'value-c') AND !empty(FieldC)
    state "Exit Condition 2" as StateB_X2 : is(FieldB, 'value-d') AND FieldD > 0
  }

  [*] --> Begin
  Begin --> StateB : FieldE == 0
@enduml
`

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
	s := &scanner{d: []byte(simpleCode)}
	for getToken(s) != nil {
		i++
	}

	a.Equal(64, i)
}

func BenchmarkTokeniser(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := &scanner{d: []byte(simpleCode)}
		for getToken(s) != nil {
		}
	}
}

func TestParser(t *testing.T) {
	a := assert.New(t)

	doc, err := parseDocument(&scanner{d: []byte(simpleCode)})
	a.NoError(err)
	a.NotNil(doc)

	a.Equal("Begin", doc.FindInitialState().Name)
	a.Equal("Value1", doc.GetSkinParam("Param1"))
	a.Equal("Value2", doc.GetSkinParam("Param2"))
	a.Equal("", doc.GetSkinParam("Param3"))
}

func BenchmarkParser(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parseDocument(&scanner{d: []byte(simpleCode)})
	}
}

func TestFormatter(t *testing.T) {
	a := assert.New(t)

	doc, err := parseDocument(&scanner{d: []byte(simpleCode)})
	a.NoError(err)
	a.NotNil(doc)

	buf := bytes.NewBuffer(nil)
	a.NoError(FormatDocument(*doc, buf))
	a.Equal(simpleCodeFormatted, buf.String())
}

func BenchmarkFormatter(b *testing.B) {
	doc, _ := parseDocument(&scanner{d: []byte(simpleCode)})

	for i := 0; i < b.N; i++ {
		FormatDocument(*doc, ioutil.Discard)
	}
}
