package formatter

import (
	"bytes"
	"fmt"

	"github.com/d5/tengo/v2/parser"
	"github.com/d5/tengo/v2/token"
)

// DefaultMaxLineLen is the default maximum line length for inlining map literals.
const DefaultMaxLineLen = 100

// Config controls formatter behaviour.
type Config struct {
	MaxLineLen int // max line length for inlining map literals; 0 disables inlining
}

type commentEntry struct {
	line int
	text string
}

// Format parses src as a Tengo source file and returns formatted output.
func Format(src []byte) ([]byte, error) {
	return FormatWithConfig(src, Config{MaxLineLen: DefaultMaxLineLen})
}

// FormatWithConfig is like Format but accepts explicit configuration.
func FormatWithConfig(src []byte, cfg Config) ([]byte, error) {
	fileSet := parser.NewFileSet()
	srcFile := fileSet.AddFile("", -1, len(src))

	comments := collectComments(srcFile, src)

	p := parser.NewParser(srcFile, src, nil)
	file, err := p.ParseFile()
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	var buf bytes.Buffer
	pr := &printer{out: &buf, srcFile: srcFile, comments: comments, maxLineLen: cfg.MaxLineLen}
	pr.printFile(file)
	return buf.Bytes(), nil
}

func collectComments(srcFile *parser.SourceFile, src []byte) []commentEntry {
	var comments []commentEntry
	s := parser.NewScanner(srcFile, src, nil, parser.ScanComments)
	for {
		tok, lit, pos := s.Scan()
		if tok == token.EOF {
			break
		}
		if tok == token.Comment {
			line := srcFile.Position(pos).Line
			comments = append(comments, commentEntry{line: line, text: lit})
		}
	}
	return comments
}
