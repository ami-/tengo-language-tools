package lsp

import (
	"encoding/json"

	"github.com/d5/tengo/v2/parser"
)

func (s *Server) handleReferences(msg RequestMessage) {
	var params ReferenceParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		s.sendResponse(*msg.ID, nil, &ResponseError{Code: CodeInvalidRequest, Message: err.Error()})
		return
	}

	s.mu.RLock()
	doc := s.docs[params.TextDocument.URI]
	s.mu.RUnlock()
	if doc == nil {
		s.sendResponse(*msg.ID, []Location{}, nil)
		return
	}

	file, srcFile, _ := parseDoc(doc.Text)
	if file == nil {
		s.sendResponse(*msg.ID, []Location{}, nil)
		return
	}

	targetPos := lspPosToParserPos(srcFile, doc.Text, params.Position)
	node := findNodeAt(file, targetPos)
	ident, ok := node.(*parser.Ident)
	if !ok {
		s.sendResponse(*msg.ID, []Location{}, nil)
		return
	}

	name := ident.Name
	var locs []Location
	for _, id := range collectIdents(file) {
		if id.Name != name {
			continue
		}
		locs = append(locs, Location{
			URI:   params.TextDocument.URI,
			Range: nodeRange(srcFile, id),
		})
	}
	if locs == nil {
		locs = []Location{}
	}
	s.sendResponse(*msg.ID, locs, nil)
}
