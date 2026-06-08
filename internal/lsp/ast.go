package lsp

import (
	"reflect"
	"strings"

	"github.com/d5/tengo/v2/parser"
)

// parseDoc parses Tengo source text and returns the AST File and SourceFile.
// The File may be non-nil even when err is non-nil (partial AST on syntax errors).
func parseDoc(text string) (*parser.File, *parser.SourceFile, error) {
	fileSet := parser.NewFileSet()
	srcFile := fileSet.AddFile("", -1, len(text))
	p := parser.NewParser(srcFile, []byte(text), nil)
	file, err := p.ParseFile()
	return file, srcFile, err
}

// lspPosToOffset converts a 0-based LSP Position to a byte offset in text.
// position.character is treated as a byte index (ASCII assumption).
func lspPosToOffset(text string, pos Position) int {
	lines := strings.SplitAfter(text, "\n")
	offset := 0
	for i, line := range lines {
		if i == pos.Line {
			return offset + pos.Character
		}
		offset += len(line)
	}
	return offset + pos.Character
}

// posToLSP converts a parser.Pos to a 0-based LSP Position.
func posToLSP(srcFile *parser.SourceFile, p parser.Pos) Position {
	sfp := srcFile.Position(p)
	line := sfp.Line - 1
	col := sfp.Column - 1
	if line < 0 {
		line = 0
	}
	if col < 0 {
		col = 0
	}
	return Position{Line: line, Character: col}
}

// lspPosToParserPos converts a 0-based LSP Position to a parser.Pos.
func lspPosToParserPos(srcFile *parser.SourceFile, text string, pos Position) parser.Pos {
	return srcFile.FileSetPos(lspPosToOffset(text, pos))
}

// nodeRange returns the LSP Range covering an AST node.
func nodeRange(srcFile *parser.SourceFile, n parser.Node) Range {
	return Range{
		Start: posToLSP(srcFile, n.Pos()),
		End:   posToLSP(srcFile, n.End()),
	}
}

// walkNode traverses the AST calling fn for each node.
// If fn returns false the subtree is not descended.
func walkNode(node parser.Node, fn func(parser.Node) bool) {
	if node == nil || isNilNode(node) {
		return
	}
	if !fn(node) {
		return
	}
	switch n := node.(type) {
	case *parser.File:
		for _, s := range n.Stmts {
			walkNode(s, fn)
		}
	case *parser.BlockStmt:
		for _, s := range n.Stmts {
			walkNode(s, fn)
		}
	case *parser.AssignStmt:
		for _, e := range n.LHS {
			walkNode(e, fn)
		}
		for _, e := range n.RHS {
			walkNode(e, fn)
		}
	case *parser.ExprStmt:
		walkNode(n.Expr, fn)
	case *parser.ReturnStmt:
		walkNode(n.Result, fn)
	case *parser.IfStmt:
		walkNode(n.Init, fn)
		walkNode(n.Cond, fn)
		walkNode(n.Body, fn)
		walkNode(n.Else, fn)
	case *parser.ForStmt:
		walkNode(n.Init, fn)
		walkNode(n.Cond, fn)
		walkNode(n.Post, fn)
		walkNode(n.Body, fn)
	case *parser.ForInStmt:
		walkNode(n.Key, fn)
		walkNode(n.Value, fn)
		walkNode(n.Iterable, fn)
		walkNode(n.Body, fn)
	case *parser.BranchStmt:
		walkNode(n.Label, fn)
	case *parser.ExportStmt:
		walkNode(n.Result, fn)
	case *parser.IncDecStmt:
		walkNode(n.Expr, fn)
	case *parser.FuncLit:
		for _, p := range n.Type.Params.List {
			walkNode(p, fn)
		}
		walkNode(n.Body, fn)
	case *parser.CallExpr:
		walkNode(n.Func, fn)
		for _, a := range n.Args {
			walkNode(a, fn)
		}
	case *parser.BinaryExpr:
		walkNode(n.LHS, fn)
		walkNode(n.RHS, fn)
	case *parser.UnaryExpr:
		walkNode(n.Expr, fn)
	case *parser.CondExpr:
		walkNode(n.Cond, fn)
		walkNode(n.True, fn)
		walkNode(n.False, fn)
	case *parser.SelectorExpr:
		walkNode(n.Expr, fn)
		// Sel is always a StringLit in Tengo (dot access desugars to map key).
		// Not recursing lets findNodeAt return the SelectorExpr itself when
		// the cursor is on the selector name, enabling cross-module definition lookup.
	case *parser.IndexExpr:
		walkNode(n.Expr, fn)
		walkNode(n.Index, fn)
	case *parser.SliceExpr:
		walkNode(n.Expr, fn)
		walkNode(n.Low, fn)
		walkNode(n.High, fn)
	case *parser.ArrayLit:
		for _, e := range n.Elements {
			walkNode(e, fn)
		}
	case *parser.MapLit:
		for _, e := range n.Elements {
			walkNode(e.Value, fn)
		}
	case *parser.ParenExpr:
		walkNode(n.Expr, fn)
	case *parser.ErrorExpr:
		walkNode(n.Expr, fn)
	case *parser.ImmutableExpr:
		walkNode(n.Expr, fn)
	// Leaves: *Ident, *IntLit, *FloatLit, *StringLit, *CharLit,
	// *BoolLit, *UndefinedLit, *ImportExpr, *BadExpr — no children.
	}
}

// findNodeAt returns the deepest AST node whose [Pos, End) range contains targetPos.
func findNodeAt(file *parser.File, targetPos parser.Pos) parser.Node {
	var best parser.Node
	walkNode(file, func(node parser.Node) bool {
		if node.Pos() <= targetPos && targetPos < node.End() {
			best = node
			return true
		}
		return false
	})
	return best
}

// collectIdents returns all Ident nodes in the AST.
func collectIdents(file *parser.File) []*parser.Ident {
	var idents []*parser.Ident
	walkNode(file, func(node parser.Node) bool {
		if id, ok := node.(*parser.Ident); ok {
			idents = append(idents, id)
		}
		return true
	})
	return idents
}

// topLevelSymbols extracts document symbols from top-level assignment statements.
func topLevelSymbols(file *parser.File, srcFile *parser.SourceFile) []DocumentSymbol {
	return stmtsToSymbols(file.Stmts, srcFile)
}

func stmtsToSymbols(stmts []parser.Stmt, srcFile *parser.SourceFile) []DocumentSymbol {
	var syms []DocumentSymbol
	for _, stmt := range stmts {
		assign, ok := stmt.(*parser.AssignStmt)
		if !ok {
			continue
		}
		for i, lhs := range assign.LHS {
			ident, ok := lhs.(*parser.Ident)
			if !ok {
				continue
			}
			kind := SymbolKindVariable
			var children []DocumentSymbol
			if i < len(assign.RHS) {
				if fn, ok := assign.RHS[i].(*parser.FuncLit); ok {
					kind = SymbolKindFunction
					children = stmtsToSymbols(fn.Body.Stmts, srcFile)
				}
			}
			syms = append(syms, DocumentSymbol{
				Name:           ident.Name,
				Kind:           kind,
				Range:          nodeRange(srcFile, assign),
				SelectionRange: nodeRange(srcFile, ident),
				Children:       children,
			})
		}
	}
	return syms
}

// funcSignature renders a FuncLit's parameter names as "func(a, b, c)".
func funcSignature(name string, fn *parser.FuncLit) string {
	var params []string
	for _, p := range fn.Type.Params.List {
		params = append(params, p.Name)
	}
	sig := "func(" + strings.Join(params, ", ") + ")"
	if name != "" {
		sig = name + " := " + sig
	}
	return sig
}

// fileHeaderComment returns consecutive // comment lines from the top of text,
// skipping any leading blank lines. Used to surface module-level doc comments.
func fileHeaderComment(text string) string {
	var comments []string
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			comments = append(comments, strings.TrimSpace(strings.TrimPrefix(trimmed, "//")))
		} else if trimmed == "" && len(comments) == 0 {
			continue // skip leading blank lines before the comment block
		} else {
			break
		}
	}
	return strings.Join(comments, "\n")
}

func isNilNode(n parser.Node) bool {
	v := reflect.ValueOf(n)
	return v.Kind() == reflect.Ptr && v.IsNil()
}
