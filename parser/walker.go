package parser

import (
	"fmt"
)

type Walker interface {
	Walk(fn func(n Node) error) error
}

func Walk(n Node, fn func(n Node) error) error {
	if err := fn(n); err != nil {
		return fmt.Errorf("Walk(%T): fn returned an error: %w", n, err)
	}

	if w, ok := n.(Walker); ok {
		if err := w.Walk(func(nn Node) error {
			return Walk(nn, fn)
		}); err != nil {
			return fmt.Errorf("Walk(%T): %w", n, err)
		}
	}

	return nil
}

type VisitType int

const (
	Enter VisitType = iota
	Exit
)

func (e VisitType) String() string {
	switch e {
	case Enter:
		return "Enter"
	case Exit:
		return "Exit"
	default:
		return "UNKNOWN"
	}
}

func Visit(n Node, fn func(v VisitType, depth int, n Node) error) error {
	if err := visit(n, 0, fn, fn); err != nil {
		return fmt.Errorf("Visit(%T): %w", n, err)
	}

	return nil
}

func visit(n Node, depth int, onEnter, onExit func(v VisitType, depth int, n Node) error) error {
	if onEnter != nil {
		if err := onEnter(Enter, depth, n); err != nil {
			return fmt.Errorf("visit(%T): fn(Enter) returned an error: %w", n, err)
		}
	}

	if w, ok := n.(Walker); ok {
		if err := w.Walk(func(nn Node) error {
			return visit(nn, depth+1, onEnter, onExit)
		}); err != nil {
			return fmt.Errorf("visit(%T): %w", n, err)
		}
	}

	if onExit != nil {
		if err := onExit(Exit, depth, n); err != nil {
			return fmt.Errorf("visit(%T): fn(Exit) returned an error: %w", n, err)
		}
	}

	return nil
}
