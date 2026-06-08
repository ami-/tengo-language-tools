package formatter

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/d5/tengo/v2/parser"
	"github.com/d5/tengo/v2/token"
)

type printer struct {
	out    *bytes.Buffer
	indent int
}

func (p *printer) write(s string) {
	p.out.WriteString(s)
}

func (p *printer) writeLine(s string) {
	p.out.WriteString(strings.Repeat("\t", p.indent))
	p.out.WriteString(s)
	p.out.WriteByte('\n')
}

func (p *printer) printFile(f *parser.File) {
	for i, stmt := range f.Stmts {
		p.printStmt(stmt)
		if i < len(f.Stmts)-1 {
			p.out.WriteByte('\n')
		}
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
		p.write("\n")

	case *parser.ExprStmt:
		p.printIndent()
		p.printExpr(s.Expr)
		p.write("\n")

	case *parser.ReturnStmt:
		p.printIndent()
		if s.Result != nil {
			p.write("return ")
			p.printExpr(s.Result)
		} else {
			p.write("return")
		}
		p.write("\n")

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
		p.write("\n")

	case *parser.ForStmt:
		p.printIndent()
		p.write("for ")
		if s.Init != nil || s.Cond != nil || s.Post != nil {
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
			p.printExpr(s.Cond)
			p.write(" ")
		}
		p.printBlock(s.Body)
		p.write("\n")

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
		p.write("\n")

	case *parser.BranchStmt:
		p.writeLine(s.Token.String())

	case *parser.IncDecStmt:
		p.printIndent()
		p.printExpr(s.Expr)
		p.write(s.Token.String())
		p.write("\n")

	case *parser.ExportStmt:
		p.printIndent()
		p.write("export ")
		p.printExpr(s.Result)
		p.write("\n")

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
		p.printStmt(stmt)
	}
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
		p.write(fmt.Sprintf("%d", e.Value))

	case *parser.FloatLit:
		p.write(fmt.Sprintf("%g", e.Value))

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
		p.write("[")
		for i, el := range e.Elements {
			if i > 0 {
				p.write(", ")
			}
			p.printExpr(el)
		}
		p.write("]")

	case *parser.MapLit:
		if len(e.Elements) == 0 {
			p.write("{}")
			return
		}
		p.write("{\n")
		p.indent++
		for _, el := range e.Elements {
			p.printIndent()
			p.write(el.Key + ": ")
			p.printExpr(el.Value)
			p.write(",\n")
		}
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
				p.write(param.Name)
			}
		}
		p.write(") ")
		p.printBlock(e.Body)

	case *parser.CallExpr:
		p.printExpr(e.Func)
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

	case *parser.BinaryExpr:
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
		p.printExpr(e.Sel)

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
