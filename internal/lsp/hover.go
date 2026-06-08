package lsp

import (
	"encoding/json"
	"strings"

	"github.com/d5/tengo/v2/parser"
)

func (s *Server) handleHover(msg RequestMessage) {
	var params HoverParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		s.sendResponse(*msg.ID, nil, &ResponseError{Code: CodeInvalidRequest, Message: err.Error()})
		return
	}

	s.mu.RLock()
	doc := s.docs[params.TextDocument.URI]
	s.mu.RUnlock()
	if doc == nil {
		s.sendResponse(*msg.ID, nil, nil)
		return
	}

	file, srcFile, _ := parseDoc(doc.Text)
	if file == nil {
		s.sendResponse(*msg.ID, nil, nil)
		return
	}

	targetPos := lspPosToParserPos(srcFile, doc.Text, params.Position)
	node := findNodeAt(file, targetPos)
	ident := identFromNode(node)
	if ident == nil {
		s.sendResponse(*msg.ID, nil, nil)
		return
	}

	content := hoverContent(file, srcFile, doc.Text, ident)
	r := nodeRange(srcFile, ident)
	s.sendResponse(*msg.ID, HoverResult{
		Contents: MarkupContent{Kind: "markdown", Value: content},
		Range:    &r,
	}, nil)
}

// identFromNode extracts the relevant Ident from a node at the cursor.
func identFromNode(node parser.Node) *parser.Ident {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	case *parser.Ident:
		return n
	case *parser.CallExpr:
		if id, ok := n.Func.(*parser.Ident); ok {
			return id
		}
	case *parser.SelectorExpr:
		if id, ok := n.Expr.(*parser.Ident); ok {
			return id
		}
	}
	return nil
}

// hoverContent returns a markdown string for the given ident.
// Scans top-level assignments to find the definition and any leading comment.
func hoverContent(file *parser.File, srcFile *parser.SourceFile, text string, ident *parser.Ident) string {
	for _, stmt := range file.Stmts {
		assign, ok := stmt.(*parser.AssignStmt)
		if !ok {
			continue
		}
		for i, lhs := range assign.LHS {
			id, ok := lhs.(*parser.Ident)
			if !ok || id.Name != ident.Name {
				continue
			}
			var sig string
			if i < len(assign.RHS) {
				if fn, ok := assign.RHS[i].(*parser.FuncLit); ok {
					sig = funcSignature(ident.Name, fn)
				}
			}
			if sig == "" {
				sig = ident.Name
			}
			stmtLine := srcFile.Position(assign.Pos()).Line - 1 // 0-based
			comment := leadingComment(text, stmtLine)
			if comment != "" {
				return "```tengo\n" + sig + "\n```\n" + comment
			}
			return "```tengo\n" + sig + "\n```"
		}
	}
	return "```tengo\n" + ident.Name + "\n```"
}

// leadingComment scans backwards from line (0-based) and returns any
// consecutive // comment lines immediately above it, joined by newlines.
func leadingComment(text string, line int) string {
	lines := strings.Split(text, "\n")
	var comments []string
	for i := line - 1; i >= 0; i-- {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "//") {
			body := strings.TrimSpace(strings.TrimPrefix(trimmed, "//"))
			comments = append([]string{body}, comments...)
		} else {
			break
		}
	}
	return strings.Join(comments, "\n")
}
