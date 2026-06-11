package formatter

import (
	"testing"
)

func TestFormat_Comments(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "leading file comment",
			input: "// header\nx := 1\n",
			want:  "// header\nx := 1\n",
		},
		{
			name:  "inline trailing comment",
			input: "x := 1 // assign\n",
			want:  "x := 1 // assign\n",
		},
		{
			// No blank line added when source had none between the statement and comment.
			name:  "comment between statements",
			input: "x := 1\n// separator\ny := 2\n",
			want:  "x := 1\n// separator\ny := 2\n",
		},
		{
			name:  "comment inside block",
			input: "if true {\n\t// inside\n\tx := 1\n}\n",
			want:  "if true {\n\t// inside\n\tx := 1\n}\n",
		},
		{
			name:  "comment at end of file",
			input: "x := 1\n// eof comment\n",
			want:  "x := 1\n// eof comment\n",
		},
		{
			name:  "comment at end of block",
			input: "if true {\n\tx := 1\n\t// end\n}\n",
			want:  "if true {\n\tx := 1\n\t// end\n}\n",
		},
		{
			name:  "pure comment file",
			input: "// just a comment\n",
			want:  "// just a comment\n",
		},
		{
			name:  "block comment",
			input: "/* block */\nx := 1\n",
			want:  "/* block */\nx := 1\n",
		},
		{
			// A multiline block comment spans multiple source lines, so the
			// blank-line preservation logic inserts a blank line after it.
			name:  "multiline block comment gets trailing blank line",
			input: "/*\n * multi\n * line\n */\nx := 1\n",
			want:  "/*\n * multi\n * line\n */\n\nx := 1\n",
		},
		{
			// Mid-expression block comments are moved to end-of-line.
			name:  "inline block comment moves to end of line",
			input: "x := /* why */ 1\n",
			want:  "x := 1 /* why */\n",
		},
		{
			name:  "multiple leading comments",
			input: "// line 1\n// line 2\nx := 1\n",
			want:  "// line 1\n// line 2\nx := 1\n",
		},
		{
			name:  "inline comment on return",
			input: "f := func() {\n\treturn 1 // done\n}\n",
			want:  "f := func() {\n\treturn 1 // done\n}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Format([]byte(tt.input))
			if err != nil {
				t.Fatalf("Format error: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("got:\n%q\nwant:\n%q", string(got), tt.want)
			}
		})
	}
}
