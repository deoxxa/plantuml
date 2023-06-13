package parser

import (
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

func TestParserComments(t *testing.T) {
  a := assert.New(t)

  doc, err := parseDocument(&scanner{d: readTestFile("simple-code-1-comments-input.uml")})
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

func TestParserPositions(t *testing.T) {
  a := assert.New(t)

  doc, err := ParseDocument(string(readTestFile("tiny-code.uml")))
  a.NoError(err)
  a.NotNil(doc)

  a.Equal(doc.GetSourcePosition().String(), "1:1")
  a.Equal(doc.GetSourceRange().String(), "1:1-25:7")

  a.Equal(&DocumentNode{
    BaseNode: BaseNode{
      SourceRange: SourceRange{
        Start: SourcePosition{Offset: 0, Line: 1, Column: 1},
        End:   SourcePosition{Offset: 312, Line: 25, Column: 7},
      },
    },
    Nodes: []Node{
      SkinParamNode{
        BaseNode: BaseNode{
          SourceRange: SourceRange{
            Start: SourcePosition{Offset: 11, Line: 3, Column: 1},
            End:   SourcePosition{Offset: 33, Line: 3, Column: 23},
          },
        },
        Name:  "ParamX",
        Value: "ValueX",
      },
      NoteNode{
        BaseNode: BaseNode{
          SourceRange: SourceRange{
            Start: SourcePosition{Offset: 36, Line: 5, Column: 1},
            End:   SourcePosition{Offset: 85, Line: 10, Column: 1},
          },
        },
        Floating: true,
        Content:  "  Line 1\n  Line 2\n  Line 3",
      },
      StateNode{
        BaseNode: BaseNode{
          SourceRange: SourceRange{
            Start: SourcePosition{Offset: 86, Line: 11, Column: 1},
            End:   SourcePosition{Offset: 163, Line: 13, Column: 1},
          },
        },
        Name:       "X_Outer",
        Label:      "x-outer",
        Stereotype: "<<sdlreceive>>",
        Children: []Node{
          StateNode{
            BaseNode: BaseNode{
              SourceRange: SourceRange{
                Start: SourcePosition{Offset: 132, Line: 12, Column: 3},
                End:   SourcePosition{Offset: 162, Line: 12, Column: 33},
              },
            },
            Name:  "X_Inner",
            Label: "x-inner",
            Text:  "X",
          },
        },
      },
      PartitionNode{
        BaseNode: BaseNode{
          SourceRange: SourceRange{
            Start: SourcePosition{Offset: 184, Line: 16, Column: 3},
            End:   SourcePosition{Offset: 301, Line: 22, Column: 7},
          },
        },
        Label: "X",
        Children: []Node{
          IfNode{
            BaseNode: BaseNode{
              SourceRange: SourceRange{
                Start: SourcePosition{Offset: 184, Line: 16, Column: 3},
                End:   SourcePosition{Offset: 301, Line: 22, Column: 7},
              },
            },
            Condition: ParenthesisNode{
              BaseNode: BaseNode{
                SourceRange: SourceRange{
                  Start: SourcePosition{Offset: 187, Line: 16, Column: 6},
                  End:   SourcePosition{Offset: 195, Line: 16, Column: 14},
                },
              },
              Content: "A == B1",
            },
            Value: ParenthesisNode{
              BaseNode: BaseNode{
                SourceRange: SourceRange{
                  Start: SourcePosition{Offset: 202, Line: 16, Column: 21},
                  End:   SourcePosition{Offset: 207, Line: 16, Column: 26},
                },
              },
              Content: "true",
            },
            Statements: []Node{
              ActionNode{
                BaseNode: BaseNode{
                  SourceRange: SourceRange{
                    Start: SourcePosition{Offset: 213, Line: 17, Column: 5},
                    End:   SourcePosition{Offset: 220, Line: 17, Column: 12},
                  },
                },
                Colour:  "Red",
                Content: "C1",
              },
            },
            Else: ElseNode{
              BaseNode: BaseNode{
                SourceRange: SourceRange{
                  Start: SourcePosition{Offset: 224, Line: 18, Column: 3},
                  End:   SourcePosition{Offset: 293, Line: 21, Column: 12},
                },
              },
              Condition: ParenthesisNode{
                BaseNode: BaseNode{
                  SourceRange: SourceRange{
                    Start: SourcePosition{Offset: 232, Line: 18, Column: 11},
                    End:   SourcePosition{Offset: 240, Line: 18, Column: 19},
                  },
                },
                Content: "A == B2",
              },
              Value: ParenthesisNode{
                BaseNode: BaseNode{
                  SourceRange: SourceRange{
                    Start: SourcePosition{Offset: 247, Line: 18, Column: 26},
                    End:   SourcePosition{Offset: 252, Line: 18, Column: 31},
                  },
                },
                Content: "true",
              },
              Statements: []Node{
                ActionNode{
                  BaseNode: BaseNode{
                    SourceRange: SourceRange{
                      Start: SourcePosition{Offset: 258, Line: 19, Column: 5},
                      End:   SourcePosition{Offset: 265, Line: 19, Column: 12},
                    },
                  },
                  Colour:  "Red",
                  Content: "C2",
                },
              },
              Else: ElseNode{
                BaseNode: BaseNode{
                  SourceRange: SourceRange{
                    Start: SourcePosition{Offset: 269, Line: 20, Column: 3},
                    End:   SourcePosition{Offset: 293, Line: 21, Column: 12},
                  },
                },
                Value: ParenthesisNode{
                  BaseNode: BaseNode{
                    SourceRange: SourceRange{
                      Start: SourcePosition{Offset: 274, Line: 20, Column: 8},
                      End:   SourcePosition{Offset: 280, Line: 20, Column: 14},
                    },
                  },
                  Content: "false",
                },
                Statements: []Node{
                  ActionNode{
                    BaseNode: BaseNode{
                      SourceRange: SourceRange{
                        Start: SourcePosition{Offset: 286, Line: 21, Column: 5},
                        End:   SourcePosition{Offset: 293, Line: 21, Column: 12},
                      },
                    },
                    Colour:  "Red",
                    Content: "C3",
                  },
                },
              },
            },
          },
        },
      },
    },
  }, doc)
}
