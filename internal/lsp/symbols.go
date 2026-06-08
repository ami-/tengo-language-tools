package lsp

import "encoding/json"

func (s *Server) handleDocumentSymbol(msg RequestMessage) {
	var params DocumentSymbolParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		s.sendResponse(*msg.ID, nil, &ResponseError{Code: CodeInvalidRequest, Message: err.Error()})
		return
	}

	s.mu.RLock()
	doc := s.docs[params.TextDocument.URI]
	s.mu.RUnlock()
	if doc == nil {
		s.sendResponse(*msg.ID, []DocumentSymbol{}, nil)
		return
	}

	file, srcFile, _ := parseDoc(doc.Text)
	syms := []DocumentSymbol{}
	if file != nil {
		syms = topLevelSymbols(file, srcFile)
	}
	s.sendResponse(*msg.ID, syms, nil)
}
