package parser

import (
	"fmt"
	"strings"
)

type tokenType int

const (
	tokenTypeNone = iota
	tokenTypeTerm
	tokenTypeLineEnd
	tokenTypeTrailing
	tokenTypeUnknown
)

func (t tokenType) String() string {
	switch t {
	case tokenTypeNone:
		return "None"
	case tokenTypeTerm:
		return "Term"
	case tokenTypeLineEnd:
		return "LineEnd"
	case tokenTypeTrailing:
		return "Trailing"
	case tokenTypeUnknown:
		return "Unknown"
	default:
		return fmt.Sprintf("INVALID<%d>", int(t))
	}
}

type token struct {
	pos int
	typ tokenType
	str string
}

func getToken(s *scanner) *token {
	s.ws()

	if s.eof() {
		return nil
	}

	if c := s.peek(); c == '\n' {
		p := s.p
		s.move(1)
		return &token{pos: p, typ: tokenTypeLineEnd}
	}

	if c := s.peek(); c == ':' {
		p := s.p

		s.move(1)
		s.ws()

		var d []byte
		for !s.eof() {
			c := s.peek()
			if c == '\n' {
				break
			}
			s.move(1)

			d = append(d, c)
		}

		return &token{pos: p, typ: tokenTypeTrailing, str: string(d)}
	}

	if c := s.peek(); c == '"' {
		p := s.p

		s.move(1)

		var d []byte
		for !s.eof() {
			c := s.byte()
			if c == '"' {
				break
			}

			d = append(d, c)
		}

		return &token{pos: p, typ: tokenTypeTerm, str: string(d)}
	}

	p := s.p

	var d []byte
	for !s.eof() {
		c := s.peek()
		if c == ' ' || c == '\t' || c == '\r' || c == '\n' {
			break
		}
		s.move(1)

		d = append(d, c)
	}

	return &token{pos: p, typ: tokenTypeTerm, str: string(d)}
}

func parseSkinParamNode(s *scanner) (*SkinParamNode, error) {
	s.savePos()

	var node SkinParamNode

	if term := getToken(s); term.str != "skinparam" {
		return nil, s.rerr(fmt.Errorf("expected skinparam term token"))
	}

	nameNode := getToken(s)
	if nameNode.typ != tokenTypeTerm {
		return nil, s.rerr(fmt.Errorf("expected term token"))
	}
	node.Name = nameNode.str

	valueNode := getToken(s)
	if valueNode.typ != tokenTypeTerm {
		return nil, s.rerr(fmt.Errorf("expected term token"))
	}
	node.Value = valueNode.str

	return &node, nil
}

func parseStateNode(s *scanner) (*StateNode, error) {
	s.savePos()

	var node StateNode

	if getToken(s).str != "state" {
		return nil, s.rerr(fmt.Errorf("expected `state'"))
	}

	nameAndLabelToken := getToken(s)
	if nameAndLabelToken.typ != tokenTypeTerm {
		return nil, s.rerr(fmt.Errorf("expected term token"))
	}

	node.Name = nameAndLabelToken.str
	node.Label = nameAndLabelToken.str

	asOrBraceOrEndToken := getToken(s)

	if asOrBraceOrEndToken.str == "as" {
		nameToken := getToken(s)
		if nameToken.typ != tokenTypeTerm {
			return nil, s.rerr(fmt.Errorf("expected term token"))
		}

		node.Name = nameToken.str

		if stereotypeToken := getToken(s); stereotypeToken.str == "<<sdlreceive>>" {
			node.Stereotype = stereotypeToken.str
			asOrBraceOrEndToken = getToken(s)
		} else {
			asOrBraceOrEndToken = stereotypeToken
		}
	}

	if asOrBraceOrEndToken.typ == tokenTypeTrailing {
		node.Text = asOrBraceOrEndToken.str
		asOrBraceOrEndToken = getToken(s)
	}

	if asOrBraceOrEndToken.typ == tokenTypeLineEnd {
		return &node, nil
	}

	if asOrBraceOrEndToken.str != "{" {
		return nil, s.rerr(fmt.Errorf("expected line end or opening brace"))
	}

	for !s.eof() {
		s.wsnl()

		tk := getToken(s)

		switch tk.str {
		case "}":
			return &node, nil
		case "---":
			node.Children = append(node.Children, SeparatorNode{})
		case "state":
			s.moveTo(tk)

			stateNode, err := parseStateNode(s)
			if err != nil {
				return nil, s.rerr(fmt.Errorf("parseStateNode: %w", err))
			}

			if stateNode != nil {
				node.Children = append(node.Children, *stateNode)
			}
		default:
			return nil, s.rerr(fmt.Errorf("parseStateNode: unhandled token typ=%s", tk.typ))
		}
	}

	return &node, nil
}

func parseEdgeNode(s *scanner) (*EdgeNode, error) {
	s.savePos()

	var node EdgeNode

	leftNode := getToken(s)
	if leftNode.typ != tokenTypeTerm {
		return nil, s.rerr(fmt.Errorf("expected term token"))
	}
	node.Left = leftNode.str

	arrowNode := getToken(s)
	if arrowNode.typ != tokenTypeTerm {
		return nil, s.rerr(fmt.Errorf("expected term token"))
	}
	if !strings.HasPrefix(arrowNode.str, "-") || !strings.HasSuffix(arrowNode.str, ">") {
		return nil, s.rerr(fmt.Errorf("expected second term to be an arrow"))
	}
	node.Direction = arrowNode.str

	rightNode := getToken(s)
	if rightNode.typ != tokenTypeTerm {
		return nil, s.rerr(fmt.Errorf("expected term token"))
	}
	node.Right = rightNode.str

	trailing := getToken(s)
	if trailing.typ != tokenTypeTrailing {
		s.moveTo(trailing)
	} else {
		node.Text = trailing.str
	}

	return &node, nil
}

func parseDocument(s *scanner) (*DocumentNode, error) {
	var doc DocumentNode

	s.wsnl()

	startToken := getToken(s)
	if startToken.str != "@startuml" {
		return nil, s.err(fmt.Errorf("parseDocument: first token should be @startuml"))
	}

loop:
	for !s.eof() {
		s.wsnl()

		tk := getToken(s)

		switch {
		case tk.str == "@enduml":
			return &doc, nil
		case tk.str == "skinparam":
			s.moveTo(tk)

			skinParamNode, err := parseSkinParamNode(s)
			if err != nil {
				return nil, err
			}

			if skinParamNode != nil {
				doc.Nodes = append(doc.Nodes, *skinParamNode)
				continue loop
			}
		case tk.str == "state":
			s.moveTo(tk)

			stateNode, err := parseStateNode(s)
			if err != nil {
				return nil, err
			}

			if stateNode != nil {
				doc.Nodes = append(doc.Nodes, *stateNode)
				continue loop
			}
		default:
			s.moveTo(tk)

			if edgeNode, err := parseEdgeNode(s); err == nil {
				doc.Nodes = append(doc.Nodes, *edgeNode)
				continue loop
			}

			return nil, s.err(fmt.Errorf("parseDocument: unhandled token typ=%s", tk.typ))
		}
	}

	return nil, fmt.Errorf("parseDocument: couldn't find @enduml token")
}

func ParseDocument(source string) (*DocumentNode, error) {
	return parseDocument(&scanner{d: []byte(source)})
}
