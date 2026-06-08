package lsp

import (
	"github.com/d5/tengo/v2/parser"
)

func parseToDiagnostics(src []byte) []Diagnostic {
	fileSet := parser.NewFileSet()
	srcFile := fileSet.AddFile("", -1, len(src))
	p := parser.NewParser(srcFile, src, nil)
	_, err := p.ParseFile()
	if err == nil {
		return nil
	}
	var diags []Diagnostic
	if list, ok := err.(parser.ErrorList); ok {
		for _, pe := range list {
			pos := sourceFilePosToPosition(pe.Pos)
			diags = append(diags, Diagnostic{
				Range:    Range{Start: pos, End: pos},
				Severity: 1,
				Message:  pe.Msg,
				Source:   "tengols",
			})
		}
	}
	return diags
}

// parser.SourceFilePos is 1-based; LSP Position is 0-based.
func sourceFilePosToPosition(p parser.SourceFilePos) Position {
	line := p.Line - 1
	col := p.Column - 1
	if line < 0 {
		line = 0
	}
	if col < 0 {
		col = 0
	}
	return Position{Line: line, Character: col}
}
