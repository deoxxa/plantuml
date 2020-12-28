package parser

import (
	"fmt"
	"io"
)

func FormatDocument(d DocumentNode, wr io.Writer) error {
	return formatNode(d, wr, "")
}

func formatNode(n Node, wr io.Writer, indent string) error {
	switch n := n.(type) {
	case SkinParamNode:
		fmt.Fprintf(wr, "%sskinparam %s %s\n", indent, n.Name, n.Value)
	case DocumentNode:
		fmt.Fprintf(wr, "%s@startuml\n", indent)

		var lastType string
		for _, c := range n.Nodes {
			t := fmt.Sprintf("%T", c)
			if t != lastType && lastType != "" {
				fmt.Fprintf(wr, "\n")
			}
			lastType = t
			formatNode(c, wr, indent+"  ")
		}

		fmt.Fprintf(wr, "%s@enduml\n", indent)
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
				if t != lastType && lastType != "" {
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
	}

	return nil
}
