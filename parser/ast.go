package parser

import (
	"fmt"
	"strings"
)

type SourcePosition struct{ Offset, Line, Column int }

func (p SourcePosition) Before(other SourcePosition) bool {
	return p.Offset < other.Offset
}

func (p SourcePosition) After(other SourcePosition) bool {
	return p.Offset > other.Offset
}

type SourceRange struct{ Start, End SourcePosition }

func MergeRanges(a []SourceRange) SourceRange {
	if len(a) == 0 {
		return SourceRange{}
	}

	var r SourceRange = a[0]

	for _, e := range a {
		if e.Start.Offset < r.Start.Offset {
			r.Start = e.Start
		}
		if e.End.Offset > r.End.Offset {
			r.End = e.End
		}
	}

	return r
}

type BaseNode struct {
	SourceRange SourceRange
}

func (b BaseNode) GetSourcePosition() SourcePosition {
	return b.SourceRange.Start
}

func (b BaseNode) GetSourceRange() SourceRange {
	return b.SourceRange
}

func (b *BaseNode) SetSourceRange(sourceRange SourceRange) {
	b.SourceRange = sourceRange
}

func (b *BaseNode) SetSourceStart(position SourcePosition) {
	b.SourceRange.Start = position
}

func (b *BaseNode) SetSourceEnd(position SourcePosition) {
	b.SourceRange.End = position
}

type Node interface {
	IsNode()
}

type DocumentNode struct {
	BaseNode
	Nodes []Node
}

func (DocumentNode) IsNode() {}

func (d DocumentNode) FindNode(fn func(n Node) bool) Node {
	for _, node := range d.Nodes {
		if fn(node) {
			return node
		}
	}

	return nil
}

func (d DocumentNode) FindStateNode(fn func(n StateNode) bool) *StateNode {
	n := d.FindNode(func(n Node) bool {
		stateNode, ok := n.(StateNode)
		if !ok {
			return false
		}

		return fn(stateNode)
	})

	if n == nil {
		return nil
	}

	stateNode, ok := n.(StateNode)
	if !ok {
		return nil
	}

	return &stateNode
}

func (d DocumentNode) FindInitialState() *StateNode {
	return d.FindStateNode(func(n StateNode) bool {
		return n.Stereotype == "<<sdlreceive>>"
	})
}

func (d DocumentNode) GetSkinParams(name string) []string {
	var a []string

	for _, node := range d.Nodes {
		skinParamNode, ok := node.(SkinParamNode)
		if !ok {
			continue
		}

		if skinParamNode.Name == name {
			a = append(a, skinParamNode.Value)
		}
	}

	return a
}

func (d DocumentNode) GetSkinParam(name string) string {
	if a := d.GetSkinParams(name); len(a) > 0 {
		return a[0]
	}

	return ""
}

type CommentNode struct {
	BaseNode
	Content string
}

func (CommentNode) IsNode() {}

type StateNode struct {
	BaseNode
	Name       string
	Label      string
	Stereotype string
	Text       string
	Children   []Node
}

func (StateNode) IsNode() {}

func (n StateNode) GetEntryConditions() []StateNode {
	return n.getChildrenWithPrefix("Entry Condition")
}

func (n StateNode) GetEntryOptions() []StateNode {
	return n.getChildrenWithPrefix("Entry Option")
}

func (n StateNode) GetExitConditions() []StateNode {
	return n.getChildrenWithPrefix("Exit Condition")
}

func (n StateNode) GetExitOptions() []StateNode {
	return n.getChildrenWithPrefix("Exit Option")
}

func (n StateNode) getChildrenWithPrefix(prefix string) []StateNode {
	var a []StateNode

	for _, node := range n.Children {
		stateNode, ok := node.(StateNode)
		if !ok {
			continue
		}

		if strings.HasPrefix(stateNode.Label, prefix) {
			a = append(a, stateNode)
		}
	}

	return a
}

type EdgeNode struct {
	BaseNode
	Left      string
	Right     string
	Direction string
	Text      string
}

func (EdgeNode) IsNode() {}

type SkinParamNode struct {
	BaseNode
	Name  string
	Value string
}

func (SkinParamNode) IsNode() {}

type SeparatorNode struct {
	BaseNode
}

func (SeparatorNode) IsNode() {}

type NoteNode struct {
	BaseNode
	Floating bool
	Position string
	Content  string
}

func (NoteNode) IsNode() {}

type PartitionNode struct {
	BaseNode
	Label    string
	Children []Node
}

func (PartitionNode) IsNode() {}

type IfNode struct {
	BaseNode
	Condition  Node
	Value      Node
	Statements []Node
	Else       Node
}

func (IfNode) IsNode() {}

type ElseNode struct {
	BaseNode
	Condition  Node
	Value      Node
	Statements []Node
	Else       Node
}

func (ElseNode) IsNode() {}

type ParenthesisNode struct {
	BaseNode
	Content string
}

func (ParenthesisNode) IsNode() {}

type ForkNode struct {
	BaseNode
	IsAgain    bool
	Statements []Node
	ForkAgain  Node
}

func (ForkNode) IsNode() {}

type ActionNode struct {
	BaseNode
	Colour  string
	Content string
}

func (ActionNode) IsNode() {}

type StartNode struct {
	BaseNode
}

func (StartNode) IsNode() {}

type EndNode struct {
	BaseNode
}

func (EndNode) IsNode() {}

func Walk(n Node, fn func(n Node) error) error {
	if err := fn(n); err != nil {
		return err
	}

	switch n := n.(type) {
	case DocumentNode:
		for _, nn := range n.Nodes {
			if err := Walk(nn, fn); err != nil {
				return err
			}
		}
	case CommentNode:
	case StateNode:
		for _, nn := range n.Children {
			if err := Walk(nn, fn); err != nil {
				return err
			}
		}
	case EdgeNode:
	case SkinParamNode:
	case SeparatorNode:
	case NoteNode:
	case PartitionNode:
		for _, nn := range n.Children {
			if err := Walk(nn, fn); err != nil {
				return err
			}
		}
	case IfNode:
		if n.Condition != nil {
			if err := Walk(n.Condition, fn); err != nil {
				return err
			}
		}
		if n.Value != nil {
			if err := Walk(n.Value, fn); err != nil {
				return err
			}
		}
		for _, nn := range n.Statements {
			if err := Walk(nn, fn); err != nil {
				return err
			}
		}
		if n.Else != nil {
			if err := Walk(n.Else, fn); err != nil {
				return err
			}
		}
	case ElseNode:
		if n.Condition != nil {
			if err := Walk(n.Condition, fn); err != nil {
				return err
			}
		}
		if n.Value != nil {
			if err := Walk(n.Value, fn); err != nil {
				return err
			}
		}
		for _, nn := range n.Statements {
			if err := Walk(nn, fn); err != nil {
				return err
			}
		}
		if n.Else != nil {
			if err := Walk(n.Else, fn); err != nil {
				return err
			}
		}
	case ParenthesisNode:
	case ForkNode:
		for _, nn := range n.Statements {
			if err := Walk(nn, fn); err != nil {
				return err
			}
		}
		if n.ForkAgain != nil {
			if err := Walk(n.ForkAgain, fn); err != nil {
				return err
			}
		}
	case ActionNode:
	case StartNode:
	case EndNode:
	default:
		return fmt.Errorf("invalid node type %T", n)
	}

	return nil
}
