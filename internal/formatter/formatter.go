package formatter

import (
	"bytes"
	"fmt"

	"github.com/d5/tengo/v2/parser"
)

// Format parses src as a Tengo source file and returns formatted output.
func Format(src []byte) ([]byte, error) {
	fileSet := parser.NewFileSet()
	srcFile := fileSet.AddFile("", -1, len(src))
	p := parser.NewParser(srcFile, src, nil)
	file, err := p.ParseFile()
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	var buf bytes.Buffer
	pr := &printer{out: &buf}
	pr.printFile(file)
	return buf.Bytes(), nil
}
