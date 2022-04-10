package parser

import (
	"fmt"
	"io"
	"strings"
)

func FormatDocument(d DocumentNode, wr io.Writer) error {
	return formatNode(d, wr, "")
}

func formatNode(n Node, wr io.Writer, indent string) error {
	switch n := n.(type) {
	case SkinParamNode:
		fmt.Fprintf(wr, "%sskinparam %s %s\n", indent, n.Name, n.Value)
	case DocumentNode:
		fmt.Fprintf(wr, "%s@startuml\n\n", indent)

		var lastType string
		for _, c := range n.Nodes {
			t := fmt.Sprintf("%T", c)
			if t != lastType && lastType != "" {
				fmt.Fprintf(wr, "\n")
			}
			lastType = t
			formatNode(c, wr, indent)
		}

		fmt.Fprintf(wr, "\n%s@enduml\n", indent)
	case CommentNode:
	case StateNode:
		if n.Name == n.Label {
			fmt.Fprintf(wr, "%sstate %s", indent, n.Name)
		} else {
			fmt.Fprintf(wr, "%sstate %q as %s", indent, n.Label, n.Name)
		}

		if len(n.Children) > 0 {
			fmt.Fprintf(wr, " {\n")

			var lastType string
			for _, c := range n.Children {
				t := fmt.Sprintf("%T", c)
				if t != lastType && lastType != "" && t != "parser.SeparatorNode" && lastType != "parser.SeparatorNode" {
					fmt.Fprintf(wr, "\n")
				}
				lastType = t
				formatNode(c, wr, indent+"  ")
			}
			fmt.Fprintf(wr, "%s}\n", indent)
		} else {
			if n.Text != "" {
				fmt.Fprintf(wr, " : %s", n.Text)
			}

			fmt.Fprintf(wr, "\n")
		}
	case EdgeNode:
		fmt.Fprintf(wr, "%s%s %s %s", indent, n.Left, n.Direction, n.Right)
		if n.Text != "" {
			fmt.Fprintf(wr, " : %s", n.Text)
		}
		fmt.Fprintf(wr, "\n")
	case SeparatorNode:
		fmt.Fprintf(wr, "%s---\n", indent)
	case NoteNode:
		fmt.Fprintf(wr, "%s", indent)
		if n.Floating {
			fmt.Fprintf(wr, "floating ")
		}
		fmt.Fprintf(wr, "note")
		if n.Position != "" {
			fmt.Fprintf(wr, " %s", n.Position)
		}
		fmt.Fprintf(wr, "\n")
		for _, l := range strings.Split(n.Content, "\n") {
			fmt.Fprintf(wr, "%s%s\n", indent, l)
		}
		fmt.Fprintf(wr, "%sendnote\n", indent)
	case PartitionNode:
		fmt.Fprintf(wr, "%spartition %q", indent, n.Label)

		if len(n.Children) > 0 {
			fmt.Fprintf(wr, " {\n")

			var lastType string
			for _, c := range n.Children {
				t := fmt.Sprintf("%T", c)
				if t != lastType && lastType != "" {
					fmt.Fprintf(wr, "\n")
				}
				lastType = t
				formatNode(c, wr, indent+"  ")
			}
			fmt.Fprintf(wr, "%s}\n", indent)
		} else {
			fmt.Fprintf(wr, " {}\n")
		}
	case IfNode:
		fmt.Fprintf(wr, "%sif", indent)
		if n.Condition != nil {
			fmt.Fprintf(wr, " ")
			formatNode(n.Condition, wr, indent)
		}
		fmt.Fprintf(wr, " then")
		if n.Value != nil {
			fmt.Fprintf(wr, " ")
			formatNode(n.Value, wr, indent)
		}
		fmt.Fprintf(wr, "\n")
		var lastType string
		for _, c := range n.Statements {
			t := fmt.Sprintf("%T", c)
			if t != lastType && lastType != "" {
				fmt.Fprintf(wr, "\n")
			}
			lastType = t
			formatNode(c, wr, indent+"  ")
		}
		if n.Else != nil {
			formatNode(n.Else, wr, indent)
		} else {
			fmt.Fprintf(wr, "%sendif\n", indent)
		}
	case ElseNode:
		fmt.Fprintf(wr, "%selse", indent)
		if n.Condition != nil {
			fmt.Fprintf(wr, " if ")
			formatNode(n.Condition, wr, indent)
			fmt.Fprintf(wr, " then")
		}
		if n.Value != nil {
			fmt.Fprintf(wr, " ")
			formatNode(n.Value, wr, indent)
		}
		fmt.Fprintf(wr, "\n")
		var lastType string
		for _, c := range n.Statements {
			t := fmt.Sprintf("%T", c)
			if t != lastType && lastType != "" {
				fmt.Fprintf(wr, "\n")
			}
			lastType = t
			formatNode(c, wr, indent+"  ")
		}
		if n.Else != nil {
			formatNode(n.Else, wr, indent)
		} else {
			fmt.Fprintf(wr, "%sendif\n", indent)
		}
	case ForkNode:
		if n.IsAgain {
			fmt.Fprintf(wr, "%sforkagain", indent)
		} else {
			fmt.Fprintf(wr, "%sfork", indent)
		}
		fmt.Fprintf(wr, "\n")
		var lastType string
		for _, c := range n.Statements {
			t := fmt.Sprintf("%T", c)
			if t != lastType && lastType != "" {
				fmt.Fprintf(wr, "\n")
			}
			lastType = t
			formatNode(c, wr, indent+"  ")
		}
		if n.ForkAgain != nil {
			formatNode(n.ForkAgain, wr, indent)
		} else {
			fmt.Fprintf(wr, "%sendfork\n", indent)
		}
	case ParenthesisNode:
		fmt.Fprintf(wr, "("+n.Content+")")
	case StartNode:
		fmt.Fprintf(wr, "%sstart\n", indent)
	case EndNode:
		fmt.Fprintf(wr, "%send\n", indent)
	case ActionNode:
		if n.Colour != "" {
			fmt.Fprintf(wr, "%s#%s:%s;\n", indent, n.Colour, n.Content)
		} else {
			fmt.Fprintf(wr, "%s:%s;\n", indent, n.Content)
		}
	default:
		fmt.Fprintf(wr, "UNRECOGNISED NODE TYPE: %T\n", n)
	}

	return nil
}
