package lsp

import (
	"encoding/json"

	"github.com/d5/tengo/v2/parser"
)

func (s *Server) handleDefinition(msg RequestMessage) {
	var params HoverParams // same shape: textDocument + position
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

	loc := findDefinition(file, srcFile, params.TextDocument.URI, ident.Name)
	s.sendResponse(*msg.ID, loc, nil) // nil if not found (client handles gracefully)
}

// findDefinition scans top-level assignments for the first LHS ident matching name.
func findDefinition(file *parser.File, srcFile *parser.SourceFile, uri, name string) *Location {
	for _, stmt := range file.Stmts {
		assign, ok := stmt.(*parser.AssignStmt)
		if !ok {
			continue
		}
		for _, lhs := range assign.LHS {
			id, ok := lhs.(*parser.Ident)
			if ok && id.Name == name {
				r := nodeRange(srcFile, id)
				return &Location{URI: uri, Range: r}
			}
		}
	}
	return nil
}
