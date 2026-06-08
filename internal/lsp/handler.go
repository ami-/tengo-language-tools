package lsp

import (
	"encoding/json"
	"os"
)

func (s *Server) dispatch(msg RequestMessage) {
	isRequest := msg.ID != nil

	if !s.initialized && msg.Method != "initialize" {
		if isRequest {
			s.sendResponse(*msg.ID, nil, &ResponseError{
				Code:    CodeServerNotInitialized,
				Message: "server not initialized",
			})
		}
		return
	}

	if s.shutdown && msg.Method != "exit" {
		if isRequest {
			s.sendResponse(*msg.ID, nil, &ResponseError{
				Code:    CodeInvalidRequest,
				Message: "shutting down",
			})
		}
		return
	}

	switch msg.Method {
	case "initialize":
		s.handleInitialize(msg)
	case "initialized":
		// notification, no reply
	case "shutdown":
		s.handleShutdown(msg)
	case "exit":
		if s.shutdown {
			os.Exit(0)
		}
		os.Exit(1)
	case "textDocument/didOpen":
		s.handleDidOpen(msg)
	case "textDocument/didChange":
		s.handleDidChange(msg)
	case "textDocument/didClose":
		s.handleDidClose(msg)
	case "textDocument/hover":
		s.handleHover(msg)
	case "textDocument/references":
		s.handleReferences(msg)
	case "textDocument/documentSymbol":
		s.handleDocumentSymbol(msg)
	default:
		if isRequest {
			s.sendResponse(*msg.ID, nil, &ResponseError{
				Code:    CodeMethodNotFound,
				Message: "method not found: " + msg.Method,
			})
		}
	}
}

func (s *Server) handleInitialize(msg RequestMessage) {
	s.initialized = true
	result := InitializeResult{
		Capabilities: ServerCapabilities{
				TextDocumentSync:       1,
				HoverProvider:          true,
				ReferencesProvider:     true,
				DocumentSymbolProvider: true,
			},
		ServerInfo:   ServerInfo{Name: "tengols", Version: s.version},
	}
	s.sendResponse(*msg.ID, result, nil)
}

func (s *Server) handleShutdown(msg RequestMessage) {
	s.shutdown = true
	s.sendResponse(*msg.ID, nil, nil)
}

func (s *Server) handleDidOpen(msg RequestMessage) {
	var params DidOpenTextDocumentParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return
	}
	td := params.TextDocument
	s.mu.Lock()
	s.docs[td.URI] = &Document{Text: td.Text, Version: td.Version}
	s.mu.Unlock()
	s.publishDiagnostics(td.URI, td.Version, []byte(td.Text))
}

func (s *Server) handleDidChange(msg RequestMessage) {
	var params DidChangeTextDocumentParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return
	}
	if len(params.ContentChanges) == 0 {
		return
	}
	uri := params.TextDocument.URI
	ver := params.TextDocument.Version
	text := params.ContentChanges[0].Text
	s.mu.Lock()
	s.docs[uri] = &Document{Text: text, Version: ver}
	s.mu.Unlock()
	s.publishDiagnostics(uri, ver, []byte(text))
}

func (s *Server) handleDidClose(msg RequestMessage) {
	var params DidCloseTextDocumentParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return
	}
	uri := params.TextDocument.URI
	s.mu.Lock()
	delete(s.docs, uri)
	s.mu.Unlock()
	s.sendNotification("textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: []Diagnostic{},
	})
}

func (s *Server) publishDiagnostics(uri string, version int, src []byte) {
	diags := parseToDiagnostics(src)
	if diags == nil {
		diags = []Diagnostic{}
	}
	s.sendNotification("textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         uri,
		Version:     &version,
		Diagnostics: diags,
	})
}
