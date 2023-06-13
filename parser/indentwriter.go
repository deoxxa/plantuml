package parser

import (
	"bytes"
	"io"
)

type indentWriterState int

const (
	stateWaitingForContent indentWriterState = iota
	stateInsideContent
)

type Indenter interface {
	Indent()
	Outdent()
}

type IndentWriter struct {
	io.Writer
	state  indentWriterState
	indent []byte
	prefix []byte
}

func NewIndentWriter(wr io.Writer, indent []byte) *IndentWriter {
	return &IndentWriter{
		Writer: wr,
		state:  stateWaitingForContent,
		indent: indent,
	}
}

func (wr *IndentWriter) Write(d []byte) (int, error) {
	var consumed int

	for _, b := range d {
		consumed++

		switch b {
		case '\n':
			wr.state = stateWaitingForContent
			if _, err := wr.Writer.Write([]byte{b}); err != nil {
				return consumed, err
			}
		case ' ', '\t':
			if wr.state == stateInsideContent {
				if _, err := wr.Writer.Write([]byte{b}); err != nil {
					return consumed, err
				}
			}
		default:
			if b == '}' || b == ')' || b == ']' {
				wr.Outdent()
			}
			if wr.state == stateWaitingForContent {
				if _, err := wr.Writer.Write(wr.prefix); err != nil {
					return consumed, err
				}
				wr.state = stateInsideContent
			}
			if _, err := wr.Writer.Write([]byte{b}); err != nil {
				return consumed, err
			}
			if b == '{' || b == '(' || b == '[' {
				wr.Indent()
			}
		}
	}

	return consumed, nil
}

func (wr *IndentWriter) Indent() {
	wr.prefix = bytes.Join([][]byte{wr.indent, wr.prefix}, []byte{})
}

func (wr *IndentWriter) Outdent() {
	wr.prefix = bytes.TrimPrefix(wr.prefix, wr.indent)
}
