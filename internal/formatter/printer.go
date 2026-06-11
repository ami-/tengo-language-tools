package formatter

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"

	"github.com/d5/tengo/v2/parser"
	"github.com/d5/tengo/v2/token"
)

type printer struct {
	out         *bytes.Buffer
	indent      int
	srcFile     *parser.SourceFile
	comments    []commentEntry
	commentIdx  int
	maxLineLen  int  // 0 = always expand map literals
	forceInline bool // render maps flat regardless of length (used inside renderMapInline)
}

func (p *printer) write(s string) {
	p.out.WriteString(s)
}

func (p *printer) writeLine(s string) {
	p.out.WriteString(strings.Repeat("\t", p.indent))
	p.out.WriteString(s)
	p.out.WriteByte('\n')
}

func (p *printer) srcLine(pos parser.Pos) int {
	if p.srcFile == nil || !pos.IsValid() {
		return 0
	}
	return p.srcFile.Position(pos).Line
}

// flushCommentsBefore emits all pending comments whose source line is before
// targetLine and returns the source line of the last emitted comment (0 if none).
func (p *printer) flushCommentsBefore(targetLine int) int {
	lastLine := 0
	for p.commentIdx < len(p.comments) && p.comments[p.commentIdx].line < targetLine {
		p.writeLine(p.comments[p.commentIdx].text)
		lastLine = p.comments[p.commentIdx].line
		p.commentIdx++
	}
	return lastLine
}

// flushRemainingComments emits all remaining pending comments.
func (p *printer) flushRemainingComments() {
	for p.commentIdx < len(p.comments) {
		p.writeLine(p.comments[p.commentIdx].text)
		p.commentIdx++
	}
}

// inlineCommentAt returns a trailing comment string (prefixed with a space) if the
// next pending comment is on the given source line, consuming it. Returns "" otherwise.
func (p *printer) inlineCommentAt(line int) string {
	if p.commentIdx < len(p.comments) && p.comments[p.commentIdx].line == line {
		text := p.comments[p.commentIdx].text
		p.commentIdx++
		return " " + text
	}
	return ""
}

func (p *printer) printFile(f *parser.File) {
	for i, stmt := range f.Stmts {
		stmtLine := p.srcLine(stmt.Pos())
		if lastComment := p.flushCommentsBefore(stmtLine); lastComment > 0 && lastComment+1 < stmtLine {
			p.out.WriteByte('\n')
		}
		p.printStmt(stmt)
		if i < len(f.Stmts)-1 {
			// Only emit a blank separator when the source itself had a gap:
			// check the first line of the next thing (comment or statement).
			nextLine := p.srcLine(f.Stmts[i+1].Pos())
			if p.commentIdx < len(p.comments) && p.comments[p.commentIdx].line < nextLine {
				nextLine = p.comments[p.commentIdx].line
			}
			if nextLine > p.stmtEndLine(stmt)+1 {
				p.out.WriteByte('\n')
			}
		}
	}
	p.flushRemainingComments()
}

// stmtEndLine returns the last source line occupied by a statement.
func (p *printer) stmtEndLine(s parser.Stmt) int {
	switch s := s.(type) {
	case *parser.IfStmt:
		if s.Else != nil {
			return p.stmtEndLine(s.Else)
		}
		return p.srcLine(s.Body.RBrace)
	case *parser.ForStmt:
		return p.srcLine(s.Body.RBrace)
	case *parser.ForInStmt:
		return p.srcLine(s.Body.RBrace)
	case *parser.BlockStmt:
		return p.srcLine(s.RBrace)
	case *parser.AssignStmt:
		if len(s.RHS) > 0 {
			return p.exprEndLine(s.RHS[len(s.RHS)-1])
		}
		return p.srcLine(s.Pos())
	case *parser.ExprStmt:
		return p.exprEndLine(s.Expr)
	default:
		return p.srcLine(s.Pos())
	}
}

// exprEndLine returns the last source line occupied by an expression.
func (p *printer) exprEndLine(e parser.Expr) int {
	switch e := e.(type) {
	case *parser.FuncLit:
		return p.srcLine(e.Body.RBrace)
	default:
		return p.srcLine(e.Pos())
	}
}

func (p *printer) printStmt(s parser.Stmt) {
	switch s := s.(type) {
	case *parser.AssignStmt:
		p.printIndent()
		for i, lhs := range s.LHS {
			if i > 0 {
				p.write(", ")
			}
			p.printExpr(lhs)
		}
		p.write(" " + s.Token.String() + " ")
		for i, rhs := range s.RHS {
			if i > 0 {
				p.write(", ")
			}
			p.printExpr(rhs)
		}
		p.write(p.inlineCommentAt(p.srcLine(s.Pos())) + "\n")

	case *parser.ExprStmt:
		p.printIndent()
		p.printExpr(s.Expr)
		p.write(p.inlineCommentAt(p.srcLine(s.Pos())) + "\n")

	case *parser.ReturnStmt:
		p.printIndent()
		if s.Result != nil {
			p.write("return ")
			p.printExpr(s.Result)
		} else {
			p.write("return")
		}
		p.write(p.inlineCommentAt(p.srcLine(s.Pos())) + "\n")

	case *parser.IfStmt:
		p.printIndent()
		p.write("if ")
		if s.Init != nil {
			p.printStmtInline(s.Init)
			p.write("; ")
		}
		p.printExpr(s.Cond)
		p.write(" ")
		p.printBlock(s.Body)
		if s.Else != nil {
			p.write(" else ")
			switch els := s.Else.(type) {
			case *parser.BlockStmt:
				p.printBlock(els)
			default:
				p.printStmt(els)
				return
			}
		}
		p.write(p.inlineCommentAt(p.srcLine(s.Pos())) + "\n")

	case *parser.ForStmt:
		p.printIndent()
		p.write("for ")
		if s.Init != nil || s.Post != nil {
			// C-style: for init; cond; post {}
			if s.Init != nil {
				p.printStmtInline(s.Init)
			}
			p.write("; ")
			if s.Cond != nil {
				p.printExpr(s.Cond)
			}
			p.write("; ")
			if s.Post != nil {
				p.printStmtInline(s.Post)
			}
			p.write(" ")
		} else if s.Cond != nil {
			// While-style: for cond {}
			p.printExpr(s.Cond)
			p.write(" ")
		}
		p.printBlock(s.Body)
		p.write(p.inlineCommentAt(p.srcLine(s.Pos())) + "\n")

	case *parser.ForInStmt:
		p.printIndent()
		p.write("for ")
		if s.Key != nil {
			p.printExpr(s.Key)
			p.write(", ")
		}
		p.printExpr(s.Value)
		p.write(" in ")
		p.printExpr(s.Iterable)
		p.write(" ")
		p.printBlock(s.Body)
		p.write(p.inlineCommentAt(p.srcLine(s.Pos())) + "\n")

	case *parser.BranchStmt:
		p.printIndent()
		p.write(s.Token.String())
		p.write(p.inlineCommentAt(p.srcLine(s.Pos())) + "\n")

	case *parser.IncDecStmt:
		p.printIndent()
		p.printExpr(s.Expr)
		p.write(s.Token.String())
		p.write(p.inlineCommentAt(p.srcLine(s.Pos())) + "\n")

	case *parser.ExportStmt:
		p.printIndent()
		p.write("export ")
		p.printExpr(s.Result)
		p.write(p.inlineCommentAt(p.srcLine(s.Pos())) + "\n")

	case *parser.EmptyStmt:
		// nothing

	case *parser.BlockStmt:
		p.printBlock(s)
		p.write("\n")
	}
}

// printStmtInline prints a statement without indentation or trailing newline (for for-init/post).
func (p *printer) printStmtInline(s parser.Stmt) {
	var buf bytes.Buffer
	inner := &printer{out: &buf, indent: 0}
	inner.printStmt(s)
	trimmed := strings.TrimRight(buf.String(), "\n")
	trimmed = strings.TrimLeft(trimmed, "\t")
	p.write(trimmed)
}

func (p *printer) printBlock(b *parser.BlockStmt) {
	p.write("{\n")
	p.indent++
	for _, stmt := range b.Stmts {
		stmtLine := p.srcLine(stmt.Pos())
		if lastComment := p.flushCommentsBefore(stmtLine); lastComment > 0 && lastComment+1 < stmtLine {
			p.out.WriteByte('\n')
		}
		p.printStmt(stmt)
	}
	p.flushCommentsBefore(p.srcLine(b.RBrace))
	p.indent--
	p.printIndent()
	p.write("}")
}

func (p *printer) printIndent() {
	p.out.WriteString(strings.Repeat("\t", p.indent))
}

func (p *printer) printExpr(e parser.Expr) {
	switch e := e.(type) {
	case *parser.Ident:
		p.write(e.Name)

	case *parser.IntLit:
		p.write(e.Literal)

	case *parser.FloatLit:
		p.write(e.Literal)

	case *parser.BoolLit:
		if e.Value {
			p.write("true")
		} else {
			p.write("false")
		}

	case *parser.UndefinedLit:
		p.write("undefined")

	case *parser.StringLit:
		p.write(`"` + escapeString(e.Value) + `"`)

	case *parser.CharLit:
		p.write(fmt.Sprintf("'%s'", string(e.Value)))

	case *parser.ArrayLit:
		if p.forceInline || len(e.Elements) == 0 || p.srcLine(e.LBrack) == p.srcLine(e.RBrack) {
			// Source was single-line (or empty, or forceInline): keep inline.
			p.write("[")
			for i, el := range e.Elements {
				if i > 0 {
					p.write(", ")
				}
				p.printExpr(el)
			}
			p.write("]")
		} else {
			// Source was multi-line: try to collapse if short enough and comment-free.
			if p.maxLineLen > 0 {
				lLine := p.srcLine(e.LBrack)
				rLine := p.srcLine(e.RBrack)
				if !p.hasCommentsInRange(lLine, rLine) {
					inline := p.renderArrayInline(e)
					if p.indent*4+len(inline) <= p.maxLineLen {
						p.write(inline)
						return
					}
				}
			}
			p.write("[\n")
			p.indent++
			for i, el := range e.Elements {
				p.flushCommentsBefore(p.srcLine(el.Pos()))
				p.printIndent()
				p.printExpr(el)
				if i < len(e.Elements)-1 {
					p.write(",")
				}
				p.write("\n")
			}
			p.flushCommentsBefore(p.srcLine(e.RBrack))
			p.indent--
			p.printIndent()
			p.write("]")
		}

	case *parser.MapLit:
		if len(e.Elements) == 0 {
			p.write("{}")
			return
		}

		// forceInline: inside renderMapInline — render flat with no comment flushing.
		if p.forceInline {
			p.write("{")
			for i, el := range e.Elements {
				if i > 0 {
					p.write(", ")
				}
				p.write(mapKey(el.Key) + ": ")
				p.printExpr(el.Value)
			}
			p.write("}")
			return
		}

		// Try single-line rendering when maxLineLen is set and the map has no
		// internal comments (comments can't survive collapsing to one line).
		if p.maxLineLen > 0 {
			lLine := p.srcLine(e.LBrace)
			rLine := p.srcLine(e.RBrace)
			if !p.hasCommentsInRange(lLine, rLine) {
				inline := p.renderMapInline(e)
				if p.indent*4+len(inline) <= p.maxLineLen {
					p.write(inline)
					return
				}
			}
		}

		// Expanded multi-line rendering.
		p.write("{\n")
		p.indent++
		for i, el := range e.Elements {
			p.flushCommentsBefore(p.srcLine(el.KeyPos))
			p.printIndent()
			p.write(mapKey(el.Key) + ": ")
			p.printExpr(el.Value)
			if i < len(e.Elements)-1 {
				p.write(",")
			}
			p.write("\n")
		}
		p.flushCommentsBefore(p.srcLine(e.RBrace))
		p.indent--
		p.printIndent()
		p.write("}")

	case *parser.FuncLit:
		p.write("func(")
		if e.Type.Params != nil {
			for i, param := range e.Type.Params.List {
				if i > 0 {
					p.write(", ")
				}
				if e.Type.Params.VarArgs && i == len(e.Type.Params.List)-1 {
					p.write("...")
				}
				p.write(param.Name)
			}
		}
		p.write(") ")
		p.printBlock(e.Body)

	case *parser.CallExpr:
		p.printExpr(e.Func)
		if !p.forceInline && len(e.Args) > 0 && p.srcLine(e.LParen) != p.srcLine(e.RParen) {
			p.write("(\n")
			p.indent++
			for i, arg := range e.Args {
				p.flushCommentsBefore(p.srcLine(arg.Pos()))
				p.printIndent()
				p.printExpr(arg)
				if i < len(e.Args)-1 {
					p.write(",")
				} else if e.Ellipsis.IsValid() {
					p.write("...")
				}
				p.write("\n")
			}
			p.flushCommentsBefore(p.srcLine(e.RParen))
			p.indent--
			p.printIndent()
			p.write(")")
		} else {
			p.write("(")
			for i, arg := range e.Args {
				if i > 0 {
					p.write(", ")
				}
				p.printExpr(arg)
			}
			if e.Ellipsis.IsValid() {
				p.write("...")
			}
			p.write(")")
		}

	case *parser.BinaryExpr:
		if p.maxLineLen > 0 && !p.forceInline &&
			(e.Token == token.LOr || e.Token == token.LAnd) {
			inline := p.renderExprInline(e)
			if p.currentLineLen()+len(inline) > p.maxLineLen {
				terms := flattenBinOp(e, e.Token)
				for i, term := range terms {
					p.printExpr(term)
					if i < len(terms)-1 {
						p.write(" " + e.Token.String() + "\n")
						p.printIndent()
						p.write("\t")
					}
				}
				return
			}
		}
		p.printExpr(e.LHS)
		p.write(" " + e.Token.String() + " ")
		p.printExpr(e.RHS)

	case *parser.UnaryExpr:
		p.write(e.Token.String())
		p.printExpr(e.Expr)

	case *parser.CondExpr:
		p.printExpr(e.Cond)
		p.write(" ? ")
		p.printExpr(e.True)
		p.write(" : ")
		p.printExpr(e.False)

	case *parser.SelectorExpr:
		p.printExpr(e.Expr)
		p.write(".")
		// Sel is stored as *StringLit in Tengo's AST; write the name without quotes.
		if sel, ok := e.Sel.(*parser.StringLit); ok {
			p.write(sel.Value)
		} else {
			p.printExpr(e.Sel)
		}

	case *parser.IndexExpr:
		p.printExpr(e.Expr)
		p.write("[")
		p.printExpr(e.Index)
		p.write("]")

	case *parser.SliceExpr:
		p.printExpr(e.Expr)
		p.write("[")
		if e.Low != nil {
			p.printExpr(e.Low)
		}
		p.write(":")
		if e.High != nil {
			p.printExpr(e.High)
		}
		p.write("]")

	case *parser.ParenExpr:
		p.write("(")
		p.printExpr(e.Expr)
		p.write(")")

	case *parser.ImportExpr:
		p.write(`import("` + e.ModuleName + `")`)

	case *parser.ErrorExpr:
		p.write("error(")
		p.printExpr(e.Expr)
		p.write(")")

	case *parser.ImmutableExpr:
		p.write("immutable(")
		p.printExpr(e.Expr)
		p.write(")")

	case *parser.BadExpr:
		p.write("/* bad expr */")
	}
}

// hasCommentsInRange reports whether any not-yet-flushed comment falls on a
// line in [fromLine, toLine] (inclusive).
func (p *printer) hasCommentsInRange(fromLine, toLine int) bool {
	for i := p.commentIdx; i < len(p.comments); i++ {
		l := p.comments[i].line
		if l > toLine {
			break
		}
		if l >= fromLine {
			return true
		}
	}
	return false
}

// currentLineLen returns the number of bytes written since the last newline.
func (p *printer) currentLineLen() int {
	buf := p.out.Bytes()
	i := bytes.LastIndexByte(buf, '\n')
	if i < 0 {
		return len(buf)
	}
	return len(buf) - i - 1
}

// renderExprInline renders an expression to a flat string for length measurement.
func (p *printer) renderExprInline(e parser.Expr) string {
	var buf bytes.Buffer
	sub := &printer{out: &buf, srcFile: p.srcFile, forceInline: true}
	sub.printExpr(e)
	return buf.String()
}

// flattenBinOp collects the leaf operands of a left-associative binary chain
// for a single operator. flattenBinOp((a||b)||c, ||) → [a, b, c].
func flattenBinOp(e parser.Expr, op token.Token) []parser.Expr {
	bin, ok := e.(*parser.BinaryExpr)
	if !ok || bin.Token != op {
		return []parser.Expr{e}
	}
	return append(flattenBinOp(bin.LHS, op), bin.RHS)
}

// renderMapInline renders e as a flat one-line string, e.g. {k1: v1, k2: v2}.
// Nested maps and arrays are also rendered flat. No comments are emitted.
func (p *printer) renderMapInline(e *parser.MapLit) string {
	var buf bytes.Buffer
	sub := &printer{out: &buf, srcFile: p.srcFile, forceInline: true}
	sub.printExpr(e)
	return buf.String()
}

// renderArrayInline renders e as a flat one-line string, e.g. [v1, v2, v3].
// Nested maps and arrays are also rendered flat. No comments are emitted.
func (p *printer) renderArrayInline(e *parser.ArrayLit) string {
	var buf bytes.Buffer
	sub := &printer{out: &buf, srcFile: p.srcFile, forceInline: true}
	sub.printExpr(e)
	return buf.String()
}

// mapKey returns the key formatted for output: unquoted if it's a valid
// identifier, double-quoted otherwise (e.g. "site-A" stays quoted).
func mapKey(k string) string {
	if k == "" {
		return `""`
	}
	for i, r := range k {
		if i == 0 && !unicode.IsLetter(r) && r != '_' {
			return `"` + escapeString(k) + `"`
		}
		if i > 0 && !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return `"` + escapeString(k) + `"`
		}
	}
	return k
}

func escapeString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	return s
}

// ensure token package is used for its String() methods
var _ = token.Add
