package lsp

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/d5/tengo/v2/parser"
)

func (s *Server) handlePrepareRename(msg RequestMessage) {
	var params PrepareRenameParams
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
		s.sendResponse(*msg.ID, nil, &ResponseError{Code: CodeInvalidRequest, Message: "parse error"})
		return
	}

	targetPos := lspPosToParserPos(srcFile, doc.Text, params.Position)
	node := findNodeAt(file, targetPos)

	switch n := node.(type) {
	case *parser.Ident:
		s.sendResponse(*msg.ID, nodeRange(srcFile, n), nil)
	case *parser.SelectorExpr:
		// Cursor is on the selector — return its range.
		r := Range{
			Start: posToLSP(srcFile, n.Sel.Pos()),
			End:   posToLSP(srcFile, n.Sel.End()),
		}
		s.sendResponse(*msg.ID, r, nil)
	default:
		s.sendResponse(*msg.ID, nil, &ResponseError{Code: CodeInvalidRequest, Message: "no renameable symbol at cursor"})
	}
}

func (s *Server) handleRename(msg RequestMessage) {
	var params RenameParams
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
		s.sendResponse(*msg.ID, nil, &ResponseError{Code: CodeInvalidRequest, Message: "parse error"})
		return
	}

	targetPos := lspPosToParserPos(srcFile, doc.Text, params.Position)
	node := findNodeAt(file, targetPos)

	switch n := node.(type) {
	case *parser.Ident:
		edit := s.renameIdent(file, srcFile, params.TextDocument.URI, n.Name, params.NewName)
		s.sendResponse(*msg.ID, edit, nil)
	case *parser.SelectorExpr:
		selLit, ok := n.Sel.(*parser.StringLit)
		if !ok {
			s.sendResponse(*msg.ID, nil, &ResponseError{Code: CodeInvalidRequest, Message: "cannot rename"})
			return
		}
		// Resolve the module that owns this selector.
		scope := resolveParent(file, srcFile, doc.Text, params.TextDocument.URI, s.rootURI, n.Expr)
		if scope == nil || scope.StdlibMod != "" || scope.File == nil {
			s.sendResponse(*msg.ID, nil, &ResponseError{Code: CodeInvalidRequest, Message: "cannot rename stdlib or unresolvable symbol"})
			return
		}
		edit := s.renameExportKey(scope.File, scope.SrcFile, scope.URI, selLit.Value, params.NewName)
		s.sendResponse(*msg.ID, edit, nil)
	default:
		s.sendResponse(*msg.ID, nil, &ResponseError{Code: CodeInvalidRequest, Message: "no renameable symbol at cursor"})
	}
}

// renameIdent renames an identifier in the current file. If the name is also
// an export key, the rename propagates across workspace consumers.
func (s *Server) renameIdent(file *parser.File, srcFile *parser.SourceFile, uri, oldName, newName string) WorkspaceEdit {
	edit := WorkspaceEdit{Changes: map[string][]TextEdit{}}

	// Rename all Ident occurrences in the current file.
	for _, id := range collectIdents(file) {
		if id.Name == oldName {
			edit.Changes[uri] = append(edit.Changes[uri], TextEdit{
				Range:   nodeRange(srcFile, id),
				NewText: newName,
			})
		}
	}

	// If the name is an export key, also rename the key string and scan workspace.
	if keyPos, ok := exportKeyPos(file, oldName); ok {
		edit.Changes[uri] = append(edit.Changes[uri], TextEdit{
			Range: Range{
				Start: posToLSP(srcFile, keyPos),
				End:   posToLSP(srcFile, parser.Pos(int(keyPos)+len(oldName))),
			},
			NewText: newName,
		})
		s.addWorkspaceConsumerEdits(edit, uri, oldName, newName)
	}

	return edit
}

// renameExportKey renames an export key in a module file and all workspace consumers.
func (s *Server) renameExportKey(modFile *parser.File, modSrcFile *parser.SourceFile, modURI, oldName, newName string) WorkspaceEdit {
	edit := WorkspaceEdit{Changes: map[string][]TextEdit{}}

	// Rename the key string in the module's export map.
	if keyPos, ok := exportKeyPos(modFile, oldName); ok {
		edit.Changes[modURI] = append(edit.Changes[modURI], TextEdit{
			Range: Range{
				Start: posToLSP(modSrcFile, keyPos),
				End:   posToLSP(modSrcFile, parser.Pos(int(keyPos)+len(oldName))),
			},
			NewText: newName,
		})
	}

	// Also rename the local binding if it matches the export key (export { foo: foo }).
	for _, id := range collectIdents(modFile) {
		if id.Name == oldName {
			edit.Changes[modURI] = append(edit.Changes[modURI], TextEdit{
				Range:   nodeRange(modSrcFile, id),
				NewText: newName,
			})
		}
	}

	s.addWorkspaceConsumerEdits(edit, modURI, oldName, newName)
	return edit
}

// addWorkspaceConsumerEdits scans all workspace .tengo files and adds edits
// for alias.oldName selectors in files that import the module at modURI.
func (s *Server) addWorkspaceConsumerEdits(edit WorkspaceEdit, modURI, oldName, newName string) {
	if s.rootURI == "" {
		return
	}
	modPath := uriToPath(modURI)
	files, _ := workspaceTengoFiles(s.rootURI)
	for _, path := range files {
		fileURI := pathToURI(path)
		if fileURI == modURI {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		text := string(data)

		// Check if open in editor — use in-memory version.
		s.mu.RLock()
		if doc, ok := s.docs[fileURI]; ok {
			text = doc.Text
		}
		s.mu.RUnlock()

		f, sf, _ := parseDoc(text)
		if f == nil {
			continue
		}

		// Find import aliases that resolve to modPath.
		aliases := importAliasesForModule(f, sf, fileURI, s.rootURI, modPath)
		if len(aliases) == 0 {
			continue
		}

		// Find all alias.oldName selectors.
		var edits []TextEdit
		for _, alias := range aliases {
			walkNode(f, func(node parser.Node) bool {
				sel, ok := node.(*parser.SelectorExpr)
				if !ok {
					return true
				}
				obj, ok := sel.Expr.(*parser.Ident)
				if !ok || obj.Name != alias {
					return true
				}
				selLit, ok := sel.Sel.(*parser.StringLit)
				if !ok || selLit.Value != oldName {
					return true
				}
				edits = append(edits, TextEdit{
					Range:   Range{Start: posToLSP(sf, sel.Sel.Pos()), End: posToLSP(sf, sel.Sel.End())},
					NewText: newName,
				})
				return true
			})
		}
		if len(edits) > 0 {
			edit.Changes[fileURI] = append(edit.Changes[fileURI], edits...)
		}
	}
}

// importAliasesForModule returns all import aliases in file f that resolve to modPath.
func importAliasesForModule(f *parser.File, _ *parser.SourceFile, fileURI, rootURI, modPath string) []string {
	var aliases []string
	for _, stmt := range f.Stmts {
		assign, ok := stmt.(*parser.AssignStmt)
		if !ok {
			continue
		}
		for i, lhs := range assign.LHS {
			id, ok := lhs.(*parser.Ident)
			if !ok {
				continue
			}
			if i >= len(assign.RHS) {
				continue
			}
			imp, ok := assign.RHS[i].(*parser.ImportExpr)
			if !ok {
				continue
			}
			resolved := resolveModulePath(imp.ModuleName, fileURI, rootURI)
			if resolved == modPath {
				aliases = append(aliases, id.Name)
			}
		}
	}
	return aliases
}

// exportKeyPos returns the source position of an export map key, if present.
func exportKeyPos(file *parser.File, name string) (parser.Pos, bool) {
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
			if elem.Key == name {
				return elem.KeyPos, true
			}
		}
	}
	return 0, false
}

// workspaceTengoFiles returns paths to all .tengo files under rootURI.
func workspaceTengoFiles(rootURI string) ([]string, error) {
	root := uriToPath(rootURI)
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if !d.IsDir() && strings.HasSuffix(path, ".tengo") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
