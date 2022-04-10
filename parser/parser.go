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
	tokenTypeLine
	tokenTypeHash
	tokenTypeColon
	tokenTypeSemi
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
	case tokenTypeLine:
		return "Line"
	case tokenTypeHash:
		return "Hash"
	case tokenTypeColon:
		return "Colon"
	case tokenTypeSemi:
		return "Semi"
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

func (t token) String() string {
	return fmt.Sprintf("token{pos: %d, typ: %s, str: %q}", t.pos, t.typ.String(), t.str)
}

type options struct {
	parseTrailing bool
}

func getToken(s *scanner, opts *options) *token {
	s.ws()

	if s.eof() {
		return nil
	}

	if c := s.peek(); c == '\n' {
		p := s.p
		s.move(1)
		return &token{pos: p, typ: tokenTypeLineEnd}
	}

	if c := s.peek(); c == '#' {
		p := s.p
		s.move(1)
		return &token{pos: p, typ: tokenTypeHash}
	}

	if c := s.peek(); c == ';' {
		p := s.p
		s.move(1)
		return &token{pos: p, typ: tokenTypeSemi}
	}

	if c := s.peek(); c == ':' {
		if opts != nil && opts.parseTrailing {
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
		} else {
			p := s.p
			s.move(1)
			return &token{pos: p, typ: tokenTypeColon}
		}
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

func readToTerminator(s *scanner, terminator byte, consume bool) (string, bool) {
	if s.eof() {
		return "", false
	}

	var d []byte
	for !s.eof() {
		c := s.peek()
		if c == terminator {
			if consume {
				s.move(1)
			}

			break
		}

		s.move(1)

		d = append(d, c)
	}

	return string(d), true
}

func parseSkinParamNode(s *scanner) (*SkinParamNode, error) {
	s.savePos()

	var node SkinParamNode

	if term := getToken(s, nil); term.str != "skinparam" {
		return nil, s.rerr(fmt.Errorf("expected skinparam term token"))
	}

	nameNode := getToken(s, nil)
	if nameNode.typ != tokenTypeTerm {
		return nil, s.rerr(fmt.Errorf("expected term token"))
	}
	node.Name = nameNode.str

	valueNode := getToken(s, nil)
	if valueNode.typ != tokenTypeTerm {
		return nil, s.rerr(fmt.Errorf("expected term token"))
	}
	node.Value = valueNode.str

	return &node, nil
}

func parseStateNode(s *scanner) (*StateNode, error) {
	s.savePos()

	var node StateNode

	if getToken(s, nil).str != "state" {
		return nil, s.rerr(fmt.Errorf("expected `state'"))
	}

	nameAndLabelToken := getToken(s, nil)
	if nameAndLabelToken.typ != tokenTypeTerm {
		return nil, s.rerr(fmt.Errorf("expected term token"))
	}

	node.Name = nameAndLabelToken.str
	node.Label = nameAndLabelToken.str

	asOrBraceOrEndToken := getToken(s, &options{parseTrailing: true})

	if asOrBraceOrEndToken.str == "as" {
		nameToken := getToken(s, nil)
		if nameToken.typ != tokenTypeTerm {
			return nil, s.rerr(fmt.Errorf("expected term token"))
		}

		node.Name = nameToken.str

		if stereotypeToken := getToken(s, &options{parseTrailing: true}); stereotypeToken.str == "<<sdlreceive>>" {
			node.Stereotype = stereotypeToken.str
			asOrBraceOrEndToken = getToken(s, &options{parseTrailing: true})
		} else {
			asOrBraceOrEndToken = stereotypeToken
		}
	}

	if asOrBraceOrEndToken.typ == tokenTypeTrailing {
		node.Text = asOrBraceOrEndToken.str
		asOrBraceOrEndToken = getToken(s, nil)
	}

	if asOrBraceOrEndToken.typ == tokenTypeLineEnd {
		return &node, nil
	}

	if asOrBraceOrEndToken.str != "{" {
		return nil, s.rerr(fmt.Errorf("expected line end or opening brace"))
	}

	for !s.eof() {
		s.wsnl()

		tk := getToken(s, nil)

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

	leftNode := getToken(s, nil)
	if leftNode == nil || leftNode.typ != tokenTypeTerm {
		return nil, s.rerr(fmt.Errorf("expected term token"))
	}
	node.Left = leftNode.str

	arrowNode := getToken(s, nil)
	if arrowNode == nil || arrowNode.typ != tokenTypeTerm {
		return nil, s.rerr(fmt.Errorf("expected term token"))
	}
	if !strings.HasPrefix(arrowNode.str, "-") || !strings.HasSuffix(arrowNode.str, ">") {
		return nil, s.rerr(fmt.Errorf("expected second term to be an arrow"))
	}
	node.Direction = arrowNode.str

	rightNode := getToken(s, nil)
	if rightNode == nil || rightNode.typ != tokenTypeTerm {
		return nil, s.rerr(fmt.Errorf("expected term token"))
	}
	node.Right = rightNode.str

	if trailing := getToken(s, &options{parseTrailing: true}); trailing != nil {
		if trailing.typ != tokenTypeTrailing {
			s.moveTo(trailing)
		} else {
			node.Text = trailing.str
		}
	}

	return &node, nil
}

func getWords(s string) []string {
	var r []string
	for _, e := range strings.Split(s, " ") {
		if e != "" {
			r = append(r, e)
		}
	}
	return r
}

func parseNoteNode(s *scanner) (*NoteNode, error) {
	s.savePos()

	var node NoteNode

	firstLine, ok := readToTerminator(s, '\n', true)
	if !ok {
		return nil, s.rerr(fmt.Errorf("unexpected eof"))
	}

	words := getWords(firstLine)

	if words[0] == "floating" {
		node.Floating = true
		words = words[1:]
	}

	if words[0] != "note" {
		return nil, s.rerr(fmt.Errorf("expected `note'; got %s", words[0]))
	}

	if len(words) > 1 {
		node.Position = words[1]
	}

	var lines []string

	for {
		line, ok := readToTerminator(s, '\n', true)
		if !ok {
			break
		}

		if strings.TrimSpace(line) == "endnote" {
			node.Content = strings.Join(lines, "\n")
			return &node, nil
		}

		lines = append(lines, line)
	}

	return nil, s.rerr(fmt.Errorf("expected endnote token"))
}

func parsePartitionNode(s *scanner) (*PartitionNode, error) {
	s.savePos()

	var node PartitionNode

	if tk := getToken(s, nil); tk == nil || tk.str != "partition" {
		return nil, s.rerr(fmt.Errorf("expected `partition'"))
	}

	labelToken := getToken(s, nil)
	if labelToken == nil || labelToken.typ != tokenTypeTerm {
		return nil, s.rerr(fmt.Errorf("expected term token"))
	}
	node.Label = labelToken.str

	braceToken := getToken(s, nil)
	if braceToken == nil || braceToken.str != "{" {
		return nil, s.rerr(fmt.Errorf("expected opening brace"))
	}

	for !s.eof() {
		s.wsnl()

		tk := getToken(s, nil)

		switch {
		case tk.str == "}":
			return &node, nil
		case tk.str == "start":
			node.Children = append(node.Children, StartNode{})
		case tk.str == "end":
			node.Children = append(node.Children, EndNode{})
		case tk.str == "floating", tk.str == "note":
			s.moveTo(tk)

			noteNode, err := parseNoteNode(s)
			if err != nil {
				return nil, err
			}

			if noteNode != nil {
				node.Children = append(node.Children, *noteNode)
			}
		case tk.str == "partition":
			s.moveTo(tk)

			partitionNode, err := parsePartitionNode(s)
			if err != nil {
				return nil, err
			}

			if partitionNode != nil {
				node.Children = append(node.Children, *partitionNode)
			}
		case tk.str == "if":
			s.moveTo(tk)

			ifNode, err := parseIfNode(s)
			if err != nil {
				return nil, err
			}

			if ifNode != nil {
				node.Children = append(node.Children, *ifNode)
			}
		case tk.typ == tokenTypeColon || tk.typ == tokenTypeHash:
			s.moveTo(tk)

			actionNode, err := parseActionNode(s)
			if err != nil {
				return nil, err
			}

			if actionNode != nil {
				node.Children = append(node.Children, *actionNode)
			}
		default:
			return nil, s.rerr(fmt.Errorf("parsePartitionNode: unhandled token %s", tk))
		}
	}

	return &node, nil
}

func parseIfNode(s *scanner) (*IfNode, error) {
	s.savePos()

	var node IfNode

	if tk := getToken(s, nil); tk == nil || tk.str != "if" {
		return nil, s.rerr(fmt.Errorf("parseIfNode: expected `if'"))
	}

	s.ws()

	condition, err := parseParenthesisNode(s)
	if err != nil {
		return nil, err
	}
	node.Condition = *condition

	s.ws()

	thenToken := getToken(s, nil)
	if thenToken == nil || thenToken.str != "then" {
		return nil, s.rerr(fmt.Errorf("parseIfNode: expected `then'"))
	}

	s.ws()

	value, err := parseParenthesisNode(s)
	if err != nil {
		return nil, err
	}
	node.Value = *value

	for !s.eof() {
		s.wsnl()

		tk := getToken(s, nil)

		switch {
		case tk.str == "endif":
			return &node, nil
		case tk.str == "else":
			s.moveTo(tk)

			elseNode, err := parseElseNode(s)
			if err != nil {
				return nil, err
			}
			node.Else = *elseNode
		case tk.str == "end":
			node.Statements = append(node.Statements, EndNode{})
		case tk.str == "floating", tk.str == "note":
			s.moveTo(tk)

			noteNode, err := parseNoteNode(s)
			if err != nil {
				return nil, err
			}

			if noteNode != nil {
				node.Statements = append(node.Statements, *noteNode)
			}
		case tk.str == "partition":
			s.moveTo(tk)

			partitionNode, err := parsePartitionNode(s)
			if err != nil {
				return nil, err
			}

			if partitionNode != nil {
				node.Statements = append(node.Statements, *partitionNode)
			}
		case tk.str == "if":
			s.moveTo(tk)

			ifNode, err := parseIfNode(s)
			if err != nil {
				return nil, err
			}

			if ifNode != nil {
				node.Statements = append(node.Statements, *ifNode)
			}
		case tk.str == "fork":
			s.moveTo(tk)

			forkNode, err := parseForkNode(s)
			if err != nil {
				return nil, err
			}

			if forkNode != nil {
				node.Statements = append(node.Statements, *forkNode)
			}
		case tk.typ == tokenTypeHash || tk.typ == tokenTypeColon:
			s.moveTo(tk)

			actionNode, err := parseActionNode(s)
			if err != nil {
				return nil, err
			}

			if actionNode != nil {
				node.Statements = append(node.Statements, *actionNode)
			}
		default:
			return nil, s.rerr(fmt.Errorf("parseIfNode: unhandled token %s", tk))
		}
	}

	return &node, nil
}

func parseElseNode(s *scanner) (*ElseNode, error) {
	s.savePos()

	var node ElseNode

	if tk := getToken(s, nil); tk == nil || tk.str != "else" {
		return nil, s.rerr(fmt.Errorf("parseElseNode: expected `else'"))
	}

	s.ws()

	if tk := getToken(s, nil); tk != nil {
		if tk.str == "if" {
			s.ws()

			condition, err := parseParenthesisNode(s)
			if err != nil {
				return nil, err
			}
			node.Condition = *condition

			s.ws()

			thenToken := getToken(s, nil)
			if thenToken == nil || thenToken.str != "then" {
				return nil, s.rerr(fmt.Errorf("parseElseNode: expected `then'"))
			}

			s.ws()
		} else {
			s.moveTo(tk)
		}
	}

	value, err := parseParenthesisNode(s)
	if err != nil {
		return nil, err
	}
	node.Value = *value

	for !s.eof() {
		s.wsnl()

		tk := getToken(s, nil)

		switch {
		case tk.str == "endif":
			s.moveTo(tk)
			return &node, nil
		case tk.str == "else":
			s.moveTo(tk)

			elseNode, err := parseElseNode(s)
			if err != nil {
				return nil, err
			}
			node.Else = *elseNode
		case tk.str == "end":
			node.Statements = append(node.Statements, EndNode{})
		case tk.str == "floating", tk.str == "note":
			s.moveTo(tk)

			noteNode, err := parseNoteNode(s)
			if err != nil {
				return nil, err
			}

			if noteNode != nil {
				node.Statements = append(node.Statements, *noteNode)
			}
		case tk.str == "partition":
			s.moveTo(tk)

			partitionNode, err := parsePartitionNode(s)
			if err != nil {
				return nil, err
			}

			if partitionNode != nil {
				node.Statements = append(node.Statements, *partitionNode)
			}
		case tk.str == "if":
			s.moveTo(tk)

			ifNode, err := parseIfNode(s)
			if err != nil {
				return nil, err
			}

			if ifNode != nil {
				node.Statements = append(node.Statements, *ifNode)
			}
		case tk.str == "fork":
			s.moveTo(tk)

			forkNode, err := parseForkNode(s)
			if err != nil {
				return nil, err
			}

			if forkNode != nil {
				node.Statements = append(node.Statements, *forkNode)
			}
		case tk.typ == tokenTypeColon || tk.typ == tokenTypeHash:
			s.moveTo(tk)

			actionNode, err := parseActionNode(s)
			if err != nil {
				return nil, err
			}

			if actionNode != nil {
				node.Statements = append(node.Statements, *actionNode)
			}
		default:
			return nil, s.rerr(fmt.Errorf("parseElseNode: unhandled token %s", tk))
		}
	}

	return &node, nil
}

func parseForkNode(s *scanner) (*ForkNode, error) {
	s.savePos()

	var node ForkNode

	if tk := getToken(s, nil); tk == nil || (tk.str != "fork" && tk.str != "forkagain") {
		return nil, s.rerr(fmt.Errorf("parseForkNode: expected `fork' or `forkagain'; got %#v", tk))
	} else {
		if tk != nil && tk.str == "forkagain" {
			node.IsAgain =  true
		}
	}

	for !s.eof() {
		s.wsnl()

		tk := getToken(s, nil)

		switch {
		case tk.str == "endfork":
			return &node, nil
		case tk.str == "forkagain":
			s.moveTo(tk)

			forkAgainNode, err := parseForkNode(s)
			if err != nil {
				return nil, err
			}
			node.ForkAgain = *forkAgainNode
			return &node, nil
		case tk.str == "end":
			node.Statements = append(node.Statements, EndNode{})
		case tk.str == "if":
			s.moveTo(tk)

			ifNode, err := parseIfNode(s)
			if err != nil {
				return nil, err
			}

			if ifNode != nil {
				node.Statements = append(node.Statements, *ifNode)
			}
		case tk.typ == tokenTypeColon || tk.typ == tokenTypeHash:
			s.moveTo(tk)

			actionNode, err := parseActionNode(s)
			if err != nil {
				return nil, err
			}

			if actionNode != nil {
				node.Statements = append(node.Statements, *actionNode)
			}
		default:
			return nil, s.rerr(fmt.Errorf("parseForkNode: unhandled token %s", tk))
		}
	}

	return &node, nil
}

func parseParenthesisNode(s *scanner) (*ParenthesisNode, error) {
	s.savePos()

	var node ParenthesisNode

	leftParenthesisCharacter := s.byte()
	if leftParenthesisCharacter != '(' {
		return nil, s.rerr(fmt.Errorf("parseParenthesisNode: expected opening parenthesis, got %#v", leftParenthesisCharacter))
	}

	var depth int
	var content []byte
	var sawClosingParenthesis bool

	for !s.eof() {
		b := s.byte()

		if b == '(' {
			depth++
		} else if b == ')' {
			if depth == 0 {
				sawClosingParenthesis = true
				break
			}
			depth--
		}

		content = append(content, b)
	}

	if !sawClosingParenthesis {
		return nil, s.rerr(fmt.Errorf("parseParenthesisNode: unexpected eof; expected to find closing parenthesis"))
	}

	node.Content = string(content)

	return &node, nil
}

func parseActionNode(s *scanner) (*ActionNode, error) {
	s.savePos()

	var node ActionNode

	startToken := getToken(s, nil)
	if startToken == nil {
		return nil, s.rerr(fmt.Errorf("parseActionNode: expected valid action start token (: or #)"))
	}

	if startToken.typ == tokenTypeHash {
		var colour []byte

		for !s.eof() {
			b := s.byte()
			if b == ':' {
				s.move(-1)
				break
			}

			colour = append(colour, b)
		}

		node.Colour = string(colour)

		startToken = getToken(s, nil)
		if startToken == nil {
			return nil, s.rerr(fmt.Errorf("parseActionNode: expected valid action start token (:) but got none"))
		}
	}

	if startToken.typ != tokenTypeColon {
		return nil, s.rerr(fmt.Errorf("parseActionNode: expected valid action start token (:); got %#v", startToken))
	}

	content, ok := readToTerminator(s, ';', false)
	if !ok {
		return nil, s.rerr(fmt.Errorf("parseActionNode: couldn't get action content"))
	}
	node.Content = content

	endToken := getToken(s, nil)
	if endToken == nil || endToken.typ != tokenTypeSemi {
		return nil, s.rerr(fmt.Errorf("parseActionNode: expected valid action end token (;) but got %s", endToken))
	}

	return &node, nil
}

func parseDocument(s *scanner) (*DocumentNode, error) {
	var doc DocumentNode

	s.wsnl()

	startToken := getToken(s, nil)
	if startToken.str != "@startuml" {
		return nil, s.err(fmt.Errorf("parseDocument: first token should be @startuml"))
	}

loop:
	for !s.eof() {
		s.wsnl()

		tk := getToken(s, nil)

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
			}
		case tk.str == "start":
			s.moveTo(tk)
			s.savePos()
			doc.Nodes = append(doc.Nodes, StartNode{})
		case tk.str == "end":
			s.moveTo(tk)
			s.savePos()
			doc.Nodes = append(doc.Nodes, EndNode{})
		case tk.str == "floating", tk.str == "note":
			s.moveTo(tk)

			noteNode, err := parseNoteNode(s)
			if err != nil {
				return nil, err
			}

			if noteNode != nil {
				doc.Nodes = append(doc.Nodes, *noteNode)
			}
		case tk.str == "partition":
			s.moveTo(tk)

			partitionNode, err := parsePartitionNode(s)
			if err != nil {
				return nil, err
			}

			if partitionNode != nil {
				doc.Nodes = append(doc.Nodes, *partitionNode)
			}
		case tk.str == "if":
			s.moveTo(tk)

			ifNode, err := parseIfNode(s)
			if err != nil {
				return nil, err
			}

			if ifNode != nil {
				doc.Nodes = append(doc.Nodes, *ifNode)
			}
		case tk.str == "state":
			s.moveTo(tk)

			stateNode, err := parseStateNode(s)
			if err != nil {
				return nil, err
			}

			if stateNode != nil {
				doc.Nodes = append(doc.Nodes, *stateNode)
			}
		default:
			s.moveTo(tk)

			if edgeNode, err := parseEdgeNode(s); err == nil {
				doc.Nodes = append(doc.Nodes, *edgeNode)
				continue loop
			}

			return nil, s.err(fmt.Errorf("parseDocument: unhandled token %s", tk))
		}
	}

	return nil, fmt.Errorf("parseDocument: couldn't find @enduml token")
}

func ParseDocument(source string) (*DocumentNode, error) {
	return parseDocument(&scanner{d: []byte(source)})
}
