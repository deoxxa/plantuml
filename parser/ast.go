package parser

import (
	"fmt"
	"strings"
)

type SourcePosition struct{ Offset, Line, Column int }

func (p SourcePosition) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

func (p SourcePosition) Before(other SourcePosition) bool {
	return p.Offset < other.Offset
}

func (p SourcePosition) After(other SourcePosition) bool {
	return p.Offset > other.Offset
}

type SourceRange struct{ Start, End SourcePosition }

func (r SourceRange) String() string {
	return fmt.Sprintf("%s-%s", r.Start.String(), r.End.String())
}

func (r *SourceRange) Expand(other SourceRange) {
	if other.Start.Offset < r.Start.Offset {
		r.Start.Offset = other.Start.Offset
		r.Start.Line = other.Start.Line
		r.Start.Column = other.Start.Column
	}
	if other.End.Offset > r.End.Offset {
		r.End.Offset = other.End.Offset
		r.End.Line = other.End.Line
		r.End.Column = other.End.Column
	}
}

func MergeRanges(a []SourceRange) SourceRange {
	if len(a) == 0 {
		return SourceRange{}
	}

	r := a[0]

	for _, e := range a {
		r.Expand(e)
	}

	return r
}

type BaseNode struct {
	SourceRange SourceRange
}

func (b BaseNode) Walk(fn func(n Node) error) error {
	return nil
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
	NodeName() string
}

type DocumentNode struct {
	BaseNode
	Nodes []Node
}

func (DocumentNode) NodeName() string { return "DocumentNode" }

func (n DocumentNode) Walk(fn func(n Node) error) error {
	for i := range n.Nodes {
		if err := fn(n.Nodes[i]); err != nil {
			return fmt.Errorf("DocumentNode.Walk: could not walk Nodes[%d]: %w", i, err)
		}
	}

	return nil
}

func (n DocumentNode) FindNode(fn func(n Node) bool) Node {
	for i := range n.Nodes {
		if fn(n.Nodes[i]) {
			return n.Nodes[i]
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

func (CommentNode) NodeName() string { return "CommentNode" }

type StateNode struct {
	BaseNode
	Name       string
	Label      string
	Stereotype string
	Text       string
	Children   []Node
}

func (StateNode) NodeName() string { return "StateNode" }

func (n StateNode) Walk(fn func(n Node) error) error {
	for i := range n.Children {
		if err := fn(n.Children[i]); err != nil {
			return fmt.Errorf("StateNode.Walk: could not walk Children[%d]: %w", i, err)
		}
	}

	return nil
}

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

func (EdgeNode) NodeName() string { return "EdgeNode" }

type SkinParamNode struct {
	BaseNode
	Name  string
	Value string
}

func (SkinParamNode) NodeName() string { return "SkinParamNode" }

type SeparatorNode struct {
	BaseNode
}

func (SeparatorNode) NodeName() string { return "SeparatorNode" }

type NoteNode struct {
	BaseNode
	Floating bool
	Position string
	Content  string
}

func (NoteNode) NodeName() string { return "NoteNode" }

type PartitionNode struct {
	BaseNode
	Label    string
	Children []Node
}

func (PartitionNode) NodeName() string { return "PartitionNode" }

func (n PartitionNode) Walk(fn func(n Node) error) error {
	for i := range n.Children {
		if err := fn(n.Children[i]); err != nil {
			return fmt.Errorf("PartitionNode.Walk: could not walk Children[%d]: %w", i, err)
		}
	}

	return nil
}

type IfNode struct {
	BaseNode
	Condition  Node
	Value      Node
	Statements []Node
	Else       Node
}

func (IfNode) NodeName() string { return "IfNode" }

func (n IfNode) Walk(fn func(n Node) error) error {
	if n.Condition != nil {
		if err := fn(n.Condition); err != nil {
			return fmt.Errorf("IfNode.Walk: could not walk Condition: %w", err)
		}
	}

	if n.Value != nil {
		if err := fn(n.Value); err != nil {
			return fmt.Errorf("IfNode.Walk: could not walk Value: %w", err)
		}
	}

	for i := range n.Statements {
		if err := fn(n.Statements[i]); err != nil {
			return fmt.Errorf("IfNode.Walk: could not walk Statements[%d]: %w", i, err)
		}
	}

	if n.Else != nil {
		if err := fn(n.Else); err != nil {
			return fmt.Errorf("IfNode.Walk: could not walk Else: %w", err)
		}
	}

	return nil
}

type ElseNode struct {
	BaseNode
	Condition  Node
	Value      Node
	Statements []Node
	Else       Node
}

func (ElseNode) NodeName() string { return "ElseNode" }

func (n ElseNode) Walk(fn func(n Node) error) error {
	if n.Condition != nil {
		if err := fn(n.Condition); err != nil {
			return fmt.Errorf("ElseNode.Walk: could not walk Condition: %w", err)
		}
	}

	if n.Value != nil {
		if err := fn(n.Value); err != nil {
			return fmt.Errorf("ElseNode.Walk: could not walk Value: %w", err)
		}
	}

	for i := range n.Statements {
		if err := fn(n.Statements[i]); err != nil {
			return fmt.Errorf("ElseNode.Walk: could not walk Statements[%d]: %w", i, err)
		}
	}

	if n.Else != nil {
		if err := fn(n.Else); err != nil {
			return fmt.Errorf("ElseNode.Walk: could not walk Else: %w", err)
		}
	}

	return nil
}

type ParenthesisNode struct {
	BaseNode
	Content string
}

func (ParenthesisNode) NodeName() string { return "ParenthesisNode" }

type ForkNode struct {
	BaseNode
	IsAgain    bool
	Statements []Node
	ForkAgain  Node
}

func (ForkNode) NodeName() string { return "ForkNode" }

func (n ForkNode) Walk(fn func(n Node) error) error {
	for i := range n.Statements {
		if err := fn(n.Statements[i]); err != nil {
			return fmt.Errorf("ForkNode.Walk: could not walk Statements[%d]: %w", i, err)
		}
	}

	if n.ForkAgain != nil {
		if err := fn(n.ForkAgain); err != nil {
			return fmt.Errorf("ForkNode.Walk: could not walk ForkAgain: %w", err)
		}
	}

	return nil
}

type ActionNode struct {
	BaseNode
	Colour  string
	Content string
}

func (ActionNode) NodeName() string { return "ActionNode" }

type StartNode struct {
	BaseNode
}

func (StartNode) NodeName() string { return "StartNode" }

type EndNode struct {
	BaseNode
}

func (EndNode) NodeName() string { return "EndNode" }
