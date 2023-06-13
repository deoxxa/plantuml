package parser

import (
  "fmt"
  "testing"

  "github.com/stretchr/testify/assert"
)

var walkAndVisitTestCases = []struct {
  filename    string
  walkResult  []string
  visitResult []string
}{
  {
    "simple-code-1-input.uml",
    []string{
      "DocumentNode",
      "SkinParamNode",
      "SkinParamNode",
      "StateNode",
      "StateNode",
      "SeparatorNode",
      "StateNode",
      "StateNode",
      "StateNode",
      "StateNode",
      "EdgeNode",
      "EdgeNode",
    },
    []string{
      "Enter 0 DocumentNode",
      "Enter 1 SkinParamNode",
      "Exit 1 SkinParamNode",
      "Enter 1 SkinParamNode",
      "Exit 1 SkinParamNode",
      "Enter 1 StateNode",
      "Enter 2 StateNode",
      "Exit 2 StateNode",
      "Enter 2 SeparatorNode",
      "Exit 2 SeparatorNode",
      "Enter 2 StateNode",
      "Exit 2 StateNode",
      "Exit 1 StateNode",
      "Enter 1 StateNode",
      "Enter 2 StateNode",
      "Exit 2 StateNode",
      "Enter 2 StateNode",
      "Exit 2 StateNode",
      "Exit 1 StateNode",
      "Enter 1 EdgeNode",
      "Exit 1 EdgeNode",
      "Enter 1 EdgeNode",
      "Exit 1 EdgeNode",
      "Exit 0 DocumentNode",
    },
  },
}

func TestWalk(t *testing.T) {
  for _, tc := range walkAndVisitTestCases {
    t.Run(tc.filename, func(t *testing.T) {
      a := assert.New(t)

      doc, err := parseDocument(&scanner{d: readTestFile(tc.filename)})
      a.NoError(err)
      a.NotNil(doc)

      var walkResult []string

      a.NoError(Walk(*doc, func(n Node) error {
        walkResult = append(walkResult, n.NodeName())
        return nil
      }))

      a.Equal(len(tc.walkResult), len(walkResult))
      a.Equal(tc.walkResult, walkResult)
    })
  }
}

func BenchmarkWalk(b *testing.B) {
  for _, tc := range walkAndVisitTestCases {
    b.Run(tc.filename, func(b *testing.B) {
      doc, _ := parseDocument(&scanner{d: readTestFile(tc.filename)})

      for i := 0; i < b.N; i++ {
        Walk(*doc, func(n Node) error { return nil })
      }
    })
  }
}

func TestVisit(t *testing.T) {
  for _, tc := range walkAndVisitTestCases {
    t.Run(tc.filename, func(t *testing.T) {
      a := assert.New(t)

      doc, err := parseDocument(&scanner{d: readTestFile(tc.filename)})
      a.NoError(err)
      a.NotNil(doc)

      var visitResult []string

      a.NoError(Visit(*doc, func(v VisitType, depth int, n Node) error {
        visitResult = append(visitResult, fmt.Sprintf("%s %d %s", v, depth, n.NodeName()))
        return nil
      }))

      a.Equal(len(tc.visitResult), len(visitResult))
      a.Equal(tc.visitResult, visitResult)
    })
  }
}

func BenchmarkVisit(b *testing.B) {
  for _, tc := range walkAndVisitTestCases {
    b.Run(tc.filename, func(b *testing.B) {
      doc, _ := parseDocument(&scanner{d: readTestFile(tc.filename)})

      for i := 0; i < b.N; i++ {
        Visit(*doc, func(v VisitType, depth int, n Node) error { return nil })
      }
    })
  }
}
