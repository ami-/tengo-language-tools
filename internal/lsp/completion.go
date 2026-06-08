package lsp

import (
	"encoding/json"
	"strings"
	"unicode"

	"github.com/d5/tengo/v2/parser"
	"github.com/d5/tengo/v2/stdlib"
	"github.com/d5/tengo/v2/token"
)

var tengoKeywords = []string{
	"if", "else", "for", "in", "return", "func",
	"import", "export", "true", "false", "undefined",
	"break", "continue",
}

func (s *Server) handleCompletion(msg RequestMessage) {
	var params CompletionParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		s.sendResponse(*msg.ID, nil, &ResponseError{Code: CodeInvalidRequest, Message: err.Error()})
		return
	}

	s.mu.RLock()
	doc := s.docs[params.TextDocument.URI]
	s.mu.RUnlock()
	if doc == nil {
		s.sendResponse(*msg.ID, []CompletionItem{}, nil)
		return
	}

	file, srcFile, _ := parseDoc(doc.Text)

	// Mode B (selector): try AST-based first, then text-based fallback.
	if file != nil {
		targetPos := lspPosToParserPos(srcFile, doc.Text, params.Position)
		node := findNodeAt(file, targetPos)
		if sel, ok := node.(*parser.SelectorExpr); ok {
			scope := resolveParent(file, srcFile, doc.Text, params.TextDocument.URI, s.rootURI, sel.Expr)
			if scope != nil {
				s.sendResponse(*msg.ID, collectScopeCompletions(scope), nil)
				return
			}
		}
	}

	// Mode B fallback: cursor is right after a dot (parse may have failed or
	// SelectorExpr.Sel is empty). Extract the object name from raw text.
	line := lineAt(doc.Text, params.Position.Line)
	if objName := wordBeforeDot(line, params.Position.Character); objName != "" {
		if file != nil {
			_, srcFile2, _ := parseDoc(doc.Text)
			synthIdent := &parser.Ident{Name: objName}
			scope := resolveParent(file, srcFile2, doc.Text, params.TextDocument.URI, s.rootURI, synthIdent)
			if scope != nil {
				s.sendResponse(*msg.ID, collectScopeCompletions(scope), nil)
				return
			}
		}
	}

	// Mode A: local identifier completions.
	var items []CompletionItem
	if file != nil {
		targetPos := lspPosToParserPos(srcFile, doc.Text, params.Position)
		items = collectLocalCompletions(file, targetPos)
	}
	s.sendResponse(*msg.ID, items, nil)
}

// collectScopeCompletions lists all members of a resolved scope.
func collectScopeCompletions(scope *SearchScope) []CompletionItem {
	if scope.StdlibMod != "" {
		return stdlibCompletions(scope.StdlibMod)
	}
	if scope.Body != nil {
		return bodyKeyCompletions(scope.Body)
	}
	if scope.File != nil {
		return exportedCompletions(scope.File)
	}
	return nil
}

func stdlibCompletions(mod string) []CompletionItem {
	if mod == "enum" {
		return enumCompletions()
	}
	m := stdlibDocs[mod]
	items := make([]CompletionItem, 0, len(m))
	for name, e := range m {
		kind := 6 // Variable
		if strings.Contains(e.Sig, "(") {
			kind = 3 // Function
		}
		items = append(items, CompletionItem{
			Label:         name,
			Kind:          kind,
			Detail:        e.Sig,
			Documentation: e.Doc,
		})
	}
	return items
}

func enumCompletions() []CompletionItem {
	src := stdlib.SourceModules["enum"]
	file, _, _ := parseDoc(src)
	if file == nil {
		return nil
	}
	var items []CompletionItem
	for _, stmt := range file.Stmts {
		exp, ok := stmt.(*parser.ExportStmt)
		if !ok {
			continue
		}
		mapLit, ok := exp.Result.(*parser.MapLit)
		if !ok {
			continue
		}
		for _, elem := range mapLit.Elements {
			detail := ""
			if fn, ok := elem.Value.(*parser.FuncLit); ok {
				detail = funcSignature(elem.Key, fn)
			}
			items = append(items, CompletionItem{Label: elem.Key, Kind: 3, Detail: detail})
		}
	}
	return items
}

func bodyKeyCompletions(body *parser.BlockStmt) []CompletionItem {
	seen := map[string]bool{}
	var items []CompletionItem
	walkNode(body, func(node parser.Node) bool {
		ret, ok := node.(*parser.ReturnStmt)
		if !ok {
			return true
		}
		mapLit, ok := ret.Result.(*parser.MapLit)
		if !ok {
			return true
		}
		for _, elem := range mapLit.Elements {
			if !seen[elem.Key] {
				seen[elem.Key] = true
				items = append(items, CompletionItem{Label: elem.Key, Kind: 6})
			}
		}
		return true
	})
	return items
}

func exportedCompletions(file *parser.File) []CompletionItem {
	var items []CompletionItem
	for _, stmt := range file.Stmts {
		exp, ok := stmt.(*parser.ExportStmt)
		if !ok {
			continue
		}
		mapLit, ok := exp.Result.(*parser.MapLit)
		if !ok {
			continue
		}
		for _, elem := range mapLit.Elements {
			detail := ""
			localName := resolveExportedIdent(file, elem.Key)
			if localName == "" {
				localName = elem.Key
			}
			if fn := findFuncLit(file, localName); fn != nil {
				detail = funcSignature(elem.Key, fn)
			}
			items = append(items, CompletionItem{Label: elem.Key, Kind: 3, Detail: detail})
		}
	}
	return items
}

// collectLocalCompletions gathers `:=` variables, function params, for-in vars,
// and keywords visible at targetPos.
func collectLocalCompletions(file *parser.File, targetPos parser.Pos) []CompletionItem {
	seen := map[string]bool{}
	var items []CompletionItem

	add := func(label string, kind int, detail string) {
		if !seen[label] {
			seen[label] = true
			items = append(items, CompletionItem{Label: label, Kind: kind, Detail: detail})
		}
	}

	walkNode(file, func(node parser.Node) bool {
		switch n := node.(type) {
		case *parser.AssignStmt:
			if n.Token != token.Define {
				return true
			}
			for i, lhs := range n.LHS {
				id, ok := lhs.(*parser.Ident)
				if !ok {
					continue
				}
				kind := 6 // Variable
				detail := ""
				if i < len(n.RHS) {
					if fn, ok := n.RHS[i].(*parser.FuncLit); ok {
						kind = 3
						detail = funcSignature(id.Name, fn)
					}
				}
				add(id.Name, kind, detail)
			}
		case *parser.FuncLit:
			if n.Body == nil || n.Body.LBrace > targetPos || targetPos > n.Body.RBrace {
				return true
			}
			if n.Type != nil && n.Type.Params != nil {
				for _, param := range n.Type.Params.List {
					add(param.Name, 6, "")
				}
			}
		case *parser.ForInStmt:
			if n.Body == nil || n.Body.LBrace > targetPos || targetPos > n.Body.RBrace {
				return true
			}
			if n.Key != nil {
				add(n.Key.Name, 6, "")
			}
			if n.Value != nil {
				add(n.Value.Name, 6, "")
			}
		}
		return true
	})

	for _, kw := range tengoKeywords {
		add(kw, 14, "")
	}

	return items
}

// wordBeforeDot returns the identifier immediately before the dot in line, where
// col is the 0-based cursor column. Returns "" if the character at col-1 is not '.'.
func wordBeforeDot(line string, col int) string {
	if col <= 0 || col > len(line) {
		return ""
	}
	// Cursor is right after '.', so line[col-1] should be '.'.
	if line[col-1] != '.' {
		return ""
	}
	before := line[:col-1]
	i := len(before)
	for i > 0 {
		r := rune(before[i-1])
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			break
		}
		i--
	}
	return before[i:]
}

// lineAt returns the 0-based line from text.
func lineAt(text string, line int) string {
	start := 0
	for i := 0; i < line; i++ {
		idx := strings.IndexByte(text[start:], '\n')
		if idx < 0 {
			return ""
		}
		start += idx + 1
	}
	end := strings.IndexByte(text[start:], '\n')
	if end < 0 {
		return text[start:]
	}
	return text[start : start+end]
}
