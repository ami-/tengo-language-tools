package lsp

import (
	"encoding/json"
	"strings"

	"github.com/ami-/tengo-language-tools/internal/formatter"
)

func (s *Server) handleFormatting(msg RequestMessage) {
	var params DocumentFormattingParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		s.sendResponse(*msg.ID, nil, &ResponseError{Code: CodeInvalidRequest, Message: err.Error()})
		return
	}

	s.mu.RLock()
	doc := s.docs[params.TextDocument.URI]
	s.mu.RUnlock()
	if doc == nil {
		s.sendResponse(*msg.ID, []TextEdit{}, nil)
		return
	}

	maxLen := s.maxLineLen
	if maxLen == 0 {
		maxLen = formatter.DefaultMaxLineLen
	}
	formatted, err := formatter.FormatWithConfig([]byte(doc.Text), formatter.Config{MaxLineLen: maxLen})
	if err != nil {
		// Parse error — don't format, return empty edits
		s.sendResponse(*msg.ID, []TextEdit{}, nil)
		return
	}

	if string(formatted) == doc.Text {
		s.sendResponse(*msg.ID, []TextEdit{}, nil)
		return
	}

	s.sendResponse(*msg.ID, []TextEdit{
		{Range: wholeFileRange(doc.Text), NewText: string(formatted)},
	}, nil)
}

// wholeFileRange returns a Range spanning the entire text content.
func wholeFileRange(text string) Range {
	lines := strings.Split(text, "\n")
	lastLine := len(lines) - 1
	return Range{
		Start: Position{Line: 0, Character: 0},
		End:   Position{Line: lastLine, Character: len(lines[lastLine])},
	}
}
