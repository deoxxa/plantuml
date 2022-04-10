package parser

import (
	"strings"
)

type Node interface {
	IsNode()
}

type DocumentNode struct {
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
	Content string
}

func (CommentNode) IsNode() {}

type StateNode struct {
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
	Left      string
	Right     string
	Direction string
	Text      string
}

func (EdgeNode) IsNode() {}

type SkinParamNode struct {
	Name  string
	Value string
}

func (SkinParamNode) IsNode() {}

type SeparatorNode struct{}

func (SeparatorNode) IsNode() {}

type NoteNode struct {
	Floating bool
	Position string
	Content  string
}

func (NoteNode) IsNode() {}

type PartitionNode struct {
	Label    string
	Children []Node
}

func (PartitionNode) IsNode() {}

type IfNode struct {
	Condition  Node
	Value      Node
	Statements []Node
	Else       Node
}

func (IfNode) IsNode() {}

type ElseNode struct {
	Condition  Node
	Value      Node
	Statements []Node
	Else       Node
}

func (ElseNode) IsNode() {}

type ParenthesisNode struct {
	Content string
}

func (ParenthesisNode) IsNode() {}

type ForkNode struct {
	IsAgain    bool
	Statements []Node
	ForkAgain  Node
}

func (ForkNode) IsNode() {}

type ActionNode struct {
	Colour  string
	Content string
}

func (ActionNode) IsNode() {}

type StartNode struct{}

func (StartNode) IsNode() {}

type EndNode struct{}

func (EndNode) IsNode() {}
