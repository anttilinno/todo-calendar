// Package tmpl provides utilities for parsing and executing markdown templates
// with {{.Placeholder}} variables.
package tmpl

import (
	"strings"
	"text/template"
	"text/template/parse"
)

// ExtractPlaceholders parses the template content and returns the unique
// {{.Field}} placeholder names in order of first appearance.
func ExtractPlaceholders(content string) ([]string, error) {
	trees, err := parse.Parse("tpl", content, "{{", "}}")
	if err != nil {
		return nil, err
	}
	tree := trees["tpl"]
	seen := make(map[string]bool)
	var names []string
	walkFields(tree.Root, seen, &names)
	return names, nil
}

// walkFields recursively walks the parse tree collecting unique FieldNode names.
func walkFields(node parse.Node, seen map[string]bool, names *[]string) {
	if node == nil {
		return
	}
	switch n := node.(type) {
	case *parse.ListNode:
		if n == nil {
			return
		}
		for _, child := range n.Nodes {
			walkFields(child, seen, names)
		}
	case *parse.ActionNode:
		if n.Pipe != nil {
			walkFields(n.Pipe, seen, names)
		}
	case *parse.PipeNode:
		for _, cmd := range n.Cmds {
			walkFields(cmd, seen, names)
		}
	case *parse.CommandNode:
		for _, arg := range n.Args {
			walkFields(arg, seen, names)
		}
	case *parse.FieldNode:
		if len(n.Ident) > 0 {
			name := n.Ident[0]
			if !seen[name] {
				seen[name] = true
				*names = append(*names, name)
			}
		}
	case *parse.IfNode:
		if n.Pipe != nil {
			walkFields(n.Pipe, seen, names)
		}
		walkFields(n.List, seen, names)
		walkFields(n.ElseList, seen, names)
	case *parse.RangeNode:
		if n.Pipe != nil {
			walkFields(n.Pipe, seen, names)
		}
		walkFields(n.List, seen, names)
		walkFields(n.ElseList, seen, names)
	case *parse.WithNode:
		if n.Pipe != nil {
			walkFields(n.Pipe, seen, names)
		}
		walkFields(n.List, seen, names)
		walkFields(n.ElseList, seen, names)
	}
}

// ExecuteTemplate parses the template content and fills placeholders with the
// provided values. Missing keys produce empty strings.
func ExecuteTemplate(content string, values map[string]string) (string, error) {
	tmpl, err := template.New("tpl").Option("missingkey=zero").Parse(content)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, values); err != nil {
		return "", err
	}
	return buf.String(), nil
}
