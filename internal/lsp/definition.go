package lsp

import (
	"os"
	"path/filepath"
	"strings"

	"encoding/json"

	"github.com/d5/tengo/v2/parser"
)

// SearchScope is what a parent expression resolves to.
// File/SrcFile/URI/Text describe where to search.
// Body, if non-nil, restricts the search to a function's return values
// (used when the parent is a call whose return value is a map).
type SearchScope struct {
	File    *parser.File
	SrcFile *parser.SourceFile
	URI     string
	Text    string            // raw source text
	Body    *parser.BlockStmt // non-nil: search return maps inside this body
}

// resolveParent traces expr backwards through top-level assignments to find
// the SearchScope for a selector expression's object.
//
//	mod.X   where mod := import("pkg")    → scope = module file
//	obj.X   where obj := mod.fn(args)     → scope = fn body in module file
//	obj.X   where obj := localFn(args)    → scope = localFn body in same file
//
// Returns nil when the expression cannot be resolved.
func resolveParent(file *parser.File, srcFile *parser.SourceFile, text, uri, rootURI string, expr parser.Expr) *SearchScope {
	ident, ok := expr.(*parser.Ident)
	if !ok {
		return nil
	}
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
			if i >= len(assign.RHS) {
				continue
			}
			switch r := assign.RHS[i].(type) {
			case *parser.ImportExpr:
				// x := import("mod")
				path := resolveModulePath(r.ModuleName, uri, rootURI)
				if path == "" {
					return nil
				}
				data, err := os.ReadFile(path)
				if err != nil {
					return nil
				}
				modFile, modSrcFile, _ := parseDoc(string(data))
				if modFile == nil {
					return nil
				}
				return &SearchScope{
					File:    modFile,
					SrcFile: modSrcFile,
					URI:     pathToURI(path),
					Text:    string(data),
				}
			case *parser.CallExpr:
				switch fn := r.Func.(type) {
				case *parser.SelectorExpr:
					// obj := mod.method(args)
					parentScope := resolveParent(file, srcFile, text, uri, rootURI, fn.Expr)
					if parentScope == nil {
						return nil
					}
					methodLit, ok := fn.Sel.(*parser.StringLit)
					if !ok {
						return nil
					}
					localName := resolveExportedIdent(parentScope.File, methodLit.Value)
					if localName == "" {
						localName = methodLit.Value
					}
					funcLit := findFuncLit(parentScope.File, localName)
					if funcLit == nil {
						return nil
					}
					return &SearchScope{
						File:    parentScope.File,
						SrcFile: parentScope.SrcFile,
						URI:     parentScope.URI,
						Text:    parentScope.Text,
						Body:    funcLit.Body,
					}
				case *parser.Ident:
					// obj := localFunc(args)
					funcLit := findFuncLit(file, fn.Name)
					if funcLit == nil {
						return nil
					}
					return &SearchScope{
						File:    file,
						SrcFile: srcFile,
						URI:     uri,
						Text:    text,
						Body:    funcLit.Body,
					}
				}
			}
		}
	}
	return nil
}

// findInScope locates the definition of name within the resolved scope.
func findInScope(scope *SearchScope, name string) *Location {
	if scope.Body != nil {
		return findKeyInFuncReturns(scope.Body, scope.SrcFile, scope.URI, name)
	}
	return findDefinitionInFile(scope.File, scope.SrcFile, scope.URI, name)
}

// findKeyInFuncReturns walks a function body and returns the location of
// keyName in the first matching "return { keyName: ... }" map literal.
func findKeyInFuncReturns(body *parser.BlockStmt, srcFile *parser.SourceFile, uri, keyName string) *Location {
	var result *Location
	walkNode(body, func(node parser.Node) bool {
		if result != nil {
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
			if elem.Key == keyName {
				result = &Location{URI: uri, Range: Range{
					Start: posToLSP(srcFile, elem.KeyPos),
					End:   posToLSP(srcFile, parser.Pos(int(elem.KeyPos)+len(elem.Key))),
				}}
				return false
			}
		}
		return true
	})
	return result
}

// resolveExportedIdent follows "export { exportedName: localIdent }" and
// returns the local ident name, or empty string if not found.
func resolveExportedIdent(file *parser.File, exportedName string) string {
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
			if elem.Key == exportedName {
				if id, ok := elem.Value.(*parser.Ident); ok {
					return id.Name
				}
			}
		}
	}
	return ""
}

// findFuncLit returns the last FuncLit bound to name at top level.
// Last assignment wins, matching Tengo's runtime semantics.
func findFuncLit(file *parser.File, name string) *parser.FuncLit {
	var result *parser.FuncLit
	for _, stmt := range file.Stmts {
		assign, ok := stmt.(*parser.AssignStmt)
		if !ok {
			continue
		}
		for i, lhs := range assign.LHS {
			if id, ok := lhs.(*parser.Ident); ok && id.Name == name {
				if i < len(assign.RHS) {
					if fn, ok := assign.RHS[i].(*parser.FuncLit); ok {
						result = fn
					}
				}
			}
		}
	}
	return result
}

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

	// Cursor is on the selector part of a dot expression (e.g. "all" in foo.all).
	// walkNode does not recurse into SelectorExpr.Sel, so findNodeAt returns the
	// SelectorExpr itself when the cursor is past the object ident.
	if sel, ok := node.(*parser.SelectorExpr); ok {
		loc := findSelectorDefinition(file, srcFile, doc.Text, params.TextDocument.URI, s.rootURI, sel)
		s.sendResponse(*msg.ID, loc, nil)
		return
	}

	ident := identFromNode(node)
	if ident == nil {
		s.sendResponse(*msg.ID, nil, nil)
		return
	}

	loc := findDefinition(file, srcFile, params.TextDocument.URI, s.rootURI, ident.Name)
	s.sendResponse(*msg.ID, loc, nil)
}

// findSelectorDefinition resolves obj.sel by tracing obj through assignments
// to find the source scope, then locating sel within it.
func findSelectorDefinition(file *parser.File, srcFile *parser.SourceFile, text, uri, rootURI string, sel *parser.SelectorExpr) *Location {
	selLit, ok := sel.Sel.(*parser.StringLit)
	if !ok {
		return nil
	}
	scope := resolveParent(file, srcFile, text, uri, rootURI, sel.Expr)
	if scope == nil {
		return nil
	}
	return findInScope(scope, selLit.Value)
}

// findDefinition scans top-level assignments for the first LHS ident matching name.
// If the RHS is an import expression the module file is resolved instead.
func findDefinition(file *parser.File, srcFile *parser.SourceFile, uri, rootURI, name string) *Location {
	for _, stmt := range file.Stmts {
		assign, ok := stmt.(*parser.AssignStmt)
		if !ok {
			continue
		}
		for i, lhs := range assign.LHS {
			id, ok := lhs.(*parser.Ident)
			if !ok || id.Name != name {
				continue
			}
			if i < len(assign.RHS) {
				if imp, ok := assign.RHS[i].(*parser.ImportExpr); ok {
					return resolveModuleLocation(imp.ModuleName, uri, rootURI)
				}
			}
			r := nodeRange(srcFile, id)
			return &Location{URI: uri, Range: r}
		}
	}
	return nil
}

// findDefinitionInFile finds name in a parsed file via two strategies:
//  1. Top-level assignment:  name := ...
//  2. Export map entry:      export { name: ... }
func findDefinitionInFile(file *parser.File, srcFile *parser.SourceFile, uri, name string) *Location {
	for _, stmt := range file.Stmts {
		switch s := stmt.(type) {
		case *parser.AssignStmt:
			for _, lhs := range s.LHS {
				if id, ok := lhs.(*parser.Ident); ok && id.Name == name {
					r := nodeRange(srcFile, id)
					return &Location{URI: uri, Range: r}
				}
			}
		case *parser.ExportStmt:
			if mapLit, ok := s.Result.(*parser.MapLit); ok {
				for _, elem := range mapLit.Elements {
					if elem.Key == name {
						keyRange := Range{
							Start: posToLSP(srcFile, elem.KeyPos),
							End:   posToLSP(srcFile, parser.Pos(int(elem.KeyPos)+len(elem.Key))),
						}
						return &Location{URI: uri, Range: keyRange}
					}
				}
			}
		}
	}
	return nil
}

var tengoStdlib = map[string]bool{
	"fmt": true, "os": true, "math": true, "text": true,
	"times": true, "rand": true, "json": true, "hex": true,
	"base64": true, "enum": true,
}

func resolveModuleLocation(moduleName, docURI, rootURI string) *Location {
	path := resolveModulePath(moduleName, docURI, rootURI)
	if path == "" {
		return nil
	}
	return &Location{URI: pathToURI(path), Range: Range{}}
}

func resolveModulePath(moduleName, docURI, rootURI string) string {
	if tengoStdlib[moduleName] {
		return ""
	}
	fileName := moduleName + ".tengo"
	if path := tryResolve(filepath.Dir(uriToPath(docURI)), fileName); path != "" {
		return path
	}
	if rootURI != "" {
		if path := tryResolve(uriToPath(rootURI), fileName); path != "" {
			return path
		}
	}
	return ""
}

func uriToPath(uri string) string {
	return strings.TrimPrefix(uri, "file://")
}

func pathToURI(path string) string {
	return "file://" + path
}

func tryResolve(dir, fileName string) string {
	path := filepath.Join(dir, fileName)
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}
