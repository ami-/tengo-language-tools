package lsp

import (
	"encoding/json"
	"os"
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

	// Cursor is on the selector name in a dot expression (e.g. "all" in mod.all).
	// findNodeAt returns the SelectorExpr because walkNode does not recurse into Sel.
	if sel, ok := node.(*parser.SelectorExpr); ok {
		selLit, ok := sel.Sel.(*parser.StringLit)
		if ok {
			scope := resolveParent(file, srcFile, doc.Text, params.TextDocument.URI, s.rootURI, sel.Expr)
			if scope != nil {
				content := hoverInScope(scope, selLit.Value)
				r := Range{
					Start: posToLSP(srcFile, sel.Sel.Pos()),
					End:   posToLSP(srcFile, sel.Sel.End()),
				}
				s.sendResponse(*msg.ID, HoverResult{
					Contents: MarkupContent{Kind: "markdown", Value: content},
					Range:    &r,
				}, nil)
				return
			}
		}
		// Fall through: show hover for the object ident.
	}

	ident := identFromNode(node)
	if ident == nil {
		s.sendResponse(*msg.ID, nil, nil)
		return
	}

	content := hoverContent(file, srcFile, doc.Text, params.TextDocument.URI, s.rootURI, ident)
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

// hoverInScope builds hover markdown for name within a resolved scope.
func hoverInScope(scope *SearchScope, name string) string {
	if scope.StdlibMod != "" {
		return hoverStdlib(scope.StdlibMod, name)
	}
	if scope.Body != nil {
		return hoverMapKey(scope.Body, name)
	}
	return hoverScopeMember(scope, name)
}

// hoverMapKey finds name as a key in a return map literal and shows its value.
func hoverMapKey(body *parser.BlockStmt, name string) string {
	var result string
	walkNode(body, func(node parser.Node) bool {
		if result != "" {
			return false
		}
		ret, ok := node.(*parser.ReturnStmt)
		if !ok {
			return true
		}
		mapLit, ok := ret.Result.(*parser.MapLit)
		if !ok {
			return true
		}
		for _, elem := range mapLit.Elements {
			if elem.Key != name {
				continue
			}
			switch v := elem.Value.(type) {
			case *parser.Ident:
				result = name + ": " + v.Name
			case *parser.CallExpr:
				if id, ok := v.Func.(*parser.Ident); ok {
					result = name + ": " + id.Name + "(...)"
				} else {
					result = name
				}
			default:
				result = name
			}
			return false
		}
		return true
	})
	if result == "" {
		result = name
	}
	return "```tengo\n" + result + "\n```"
}

// hoverScopeMember shows the signature and leading comment for a member of a
// file scope — following the export map to the local function if needed.
func hoverScopeMember(scope *SearchScope, name string) string {
	localName := resolveExportedIdent(scope.File, name)
	if localName == "" {
		localName = name
	}
	for _, stmt := range scope.File.Stmts {
		assign, ok := stmt.(*parser.AssignStmt)
		if !ok {
			continue
		}
		for i, lhs := range assign.LHS {
			id, ok := lhs.(*parser.Ident)
			if !ok || id.Name != localName {
				continue
			}
			if i < len(assign.RHS) {
				if fn, ok := assign.RHS[i].(*parser.FuncLit); ok {
					sig := funcSignature(name, fn)
					if scope.Text != "" {
						stmtLine := scope.SrcFile.Position(assign.Pos()).Line - 1
						if cmt := leadingComment(scope.Text, stmtLine); cmt != "" {
							return "```tengo\n" + sig + "\n```\n" + cmt
						}
					}
					return "```tengo\n" + sig + "\n```"
				}
			}
			return "```tengo\n" + name + "\n```"
		}
	}
	return "```tengo\n" + name + "\n```"
}

// hoverContent returns a markdown string for the given ident.
// Scans top-level assignments to find the definition, any leading comment,
// and (for import bindings) the module file's header comment.
func hoverContent(file *parser.File, srcFile *parser.SourceFile, text, docURI, rootURI string, ident *parser.Ident) string {
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

			stmtLine := srcFile.Position(assign.Pos()).Line - 1
			localComment := leadingComment(text, stmtLine)

			if i < len(assign.RHS) {
				// Function literal: show signature
				if fn, ok := assign.RHS[i].(*parser.FuncLit); ok {
					sig := funcSignature(ident.Name, fn)
					if localComment != "" {
						return "```tengo\n" + sig + "\n```\n" + localComment
					}
					return "```tengo\n" + sig + "\n```"
				}
				// Import: show import sig + local comment + module header
				if imp, ok := assign.RHS[i].(*parser.ImportExpr); ok {
					sig := ident.Name + ` := import("` + imp.ModuleName + `")`
					return buildImportHover(sig, localComment, imp.ModuleName, docURI, rootURI)
				}
			}

			if localComment != "" {
				return "```tengo\n" + ident.Name + "\n```\n" + localComment
			}
			return "```tengo\n" + ident.Name + "\n```"
		}
	}
	return "```tengo\n" + ident.Name + "\n```"
}

func buildImportHover(sig, localComment, moduleName, docURI, rootURI string) string {
	var parts []string
	parts = append(parts, "```tengo\n"+sig+"\n```")
	if localComment != "" {
		parts = append(parts, localComment)
	}
	if modHeader := moduleFileHeader(moduleName, docURI, rootURI); modHeader != "" {
		parts = append(parts, "---\n"+modHeader)
	}
	return strings.Join(parts, "\n")
}

func moduleFileHeader(moduleName, docURI, rootURI string) string {
	path := resolveModulePath(moduleName, docURI, rootURI)
	if path == "" {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return fileHeaderComment(string(data))
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
