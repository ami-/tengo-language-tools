package formatter

import (
	"testing"
)

// fmtDefault formats with the default config (MaxLineLen=100).
func fmtDefault(src string) (string, error) {
	out, err := FormatWithConfig([]byte(src), Config{MaxLineLen: DefaultMaxLineLen})
	return string(out), err
}

// fmtExpand formats with MaxLineLen=0 (always expand maps).
func fmtExpand(src string) (string, error) {
	out, err := FormatWithConfig([]byte(src), Config{MaxLineLen: 0})
	return string(out), err
}

func runCases(t *testing.T, format func(string) (string, error), cases []struct{ name, input, want string }) {
	t.Helper()
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := format(tt.input)
			if err != nil {
				t.Fatalf("Format error: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("got:\n%s\nwant:\n%s", string(got), tt.want)
			}
		})
	}
}

// TestFormat_Numbers verifies that numeric literals are preserved verbatim:
// integers stay integers, floats keep their decimal point and exact text.
func TestFormat_Numbers(t *testing.T) {
	cases := []struct{ name, input, want string }{
		{
			name:  "integer literal preserved",
			input: "x := 42\n",
			want:  "x := 42\n",
		},
		{
			name:  "hex integer preserved",
			input: "x := 0xFF\n",
			want:  "x := 0xFF\n",
		},
		{
			name:  "zero float preserved",
			input: "x := 0.0\n",
			want:  "x := 0.0\n",
		},
		{
			name:  "float with trailing zero preserved",
			input: "x := 100.0\n",
			want:  "x := 100.0\n",
		},
		{
			name:  "float with significant decimals preserved",
			input: "x := 3.14\n",
			want:  "x := 3.14\n",
		},
		{
			name:  "large float preserved",
			input: "x := 150000.0\n",
			want:  "x := 150000.0\n",
		},
		{
			name:  "float in map value preserved",
			input: "m := {a: 0.0, b: 1.5, c: 200000.0}\n",
			want:  "m := {a: 0.0, b: 1.5, c: 200000.0}\n",
		},
		{
			name:  "int in map value preserved",
			input: "m := {a: 1, b: 42}\n",
			want:  "m := {a: 1, b: 42}\n",
		},
	}
	runCases(t, fmtDefault, cases)
}

// TestFormat_SelectorExpr verifies that dot selectors are not quoted.
func TestFormat_SelectorExpr(t *testing.T) {
	cases := []struct{ name, input, want string }{
		{
			name:  "simple selector",
			input: "x := foo.bar\n",
			want:  "x := foo.bar\n",
		},
		{
			name:  "chained selector",
			input: "x := a.b.c\n",
			want:  "x := a.b.c\n",
		},
		{
			name:  "selector in expression",
			input: "result := req.poi\n",
			want:  "result := req.poi\n",
		},
		{
			name:  "selector as call receiver",
			input: "lib.foo(x)\n",
			want:  "lib.foo(x)\n",
		},
	}
	runCases(t, fmtDefault, cases)
}

// TestFormat_MapKeys verifies that map keys are unquoted when they are valid
// identifiers and quoted (with double quotes) when they are not.
func TestFormat_MapKeys(t *testing.T) {
	cases := []struct{ name, input, want string }{
		{
			name:  "identifier key stays unquoted",
			input: `m := {foo: 1}` + "\n",
			want:  `m := {foo: 1}` + "\n",
		},
		{
			name:  "quoted identifier key becomes unquoted",
			input: `m := {"foo": 1}` + "\n",
			want:  `m := {foo: 1}` + "\n",
		},
		{
			name:  "hyphenated key stays quoted",
			input: `m := {"site-A": 1}` + "\n",
			want:  `m := {"site-A": 1}` + "\n",
		},
		{
			name:  "key with dot stays quoted",
			input: `m := {"a.b": 1}` + "\n",
			want:  `m := {"a.b": 1}` + "\n",
		},
		{
			name:  "underscore key is valid identifier",
			input: `m := {_key: 1}` + "\n",
			want:  `m := {_key: 1}` + "\n",
		},
		{
			// Both keys fit inline; "site-A" must stay quoted in the inline form.
			name:  "mixed valid and invalid keys",
			input: "m := {\n\tfoo: 1,\n\t\"site-A\": 2\n}\n",
			want:  "m := {foo: 1, \"site-A\": 2}\n",
		},
	}
	runCases(t, fmtDefault, cases)
}

// TestFormat_BlankLines verifies that blank lines between top-level statements
// are preserved when they exist in the source and not added when absent.
func TestFormat_BlankLines(t *testing.T) {
	cases := []struct{ name, input, want string }{
		{
			name:  "consecutive statements get no blank line",
			input: "a := 1\nb := 2\nc := 3\n",
			want:  "a := 1\nb := 2\nc := 3\n",
		},
		{
			name:  "blank line between statements is preserved",
			input: "a := 1\n\nb := 2\n",
			want:  "a := 1\n\nb := 2\n",
		},
		{
			name:  "blank line before comment is preserved",
			input: "a := 1\n\n// note\nb := 2\n",
			want:  "a := 1\n\n// note\nb := 2\n",
		},
		{
			name:  "no extra blank line between simple print calls",
			input: "fmt.println(\"a\")\nfmt.println(\"b\")\nfmt.println(\"c\")\n",
			want:  "fmt.println(\"a\")\nfmt.println(\"b\")\nfmt.println(\"c\")\n",
		},
		{
			// Blank line between comment and following statement must be preserved.
			name:  "blank line after comment before statement preserved",
			input: "// comment\n\nx := 1\n",
			want:  "// comment\n\nx := 1\n",
		},
		{
			name:  "no blank line after comment when absent",
			input: "// comment\nx := 1\n",
			want:  "// comment\nx := 1\n",
		},
		{
			name:  "blank line after comment mid-file preserved",
			input: "a := 1\n// note\n\nb := 2\n",
			want:  "a := 1\n// note\n\nb := 2\n",
		},
		{
			// Blank line after comment inside a function body.
			name:  "blank line after comment in block preserved",
			input: "f := func() {\n\t// comment\n\n\tx := 1\n}\n",
			want:  "f := func() {\n\t// comment\n\n\tx := 1\n}\n",
		},
	}
	runCases(t, fmtDefault, cases)
}

// TestFormat_CommentsInMapLit verifies that comments inside map literals are
// emitted before the element they precede, not displaced to the end.
func TestFormat_CommentsInMapLit(t *testing.T) {
	cases := []struct{ name, input, want string }{
		{
			name:  "comment before first element",
			input: "m := {\n\t// first\n\ta: 1,\n\tb: 2\n}\n",
			want:  "m := {\n\t// first\n\ta: 1,\n\tb: 2\n}\n",
		},
		{
			name:  "comment between elements",
			input: "m := {\n\ta: 1,\n\t// between\n\tb: 2\n}\n",
			want:  "m := {\n\ta: 1,\n\t// between\n\tb: 2\n}\n",
		},
		{
			name:  "comment after last element",
			input: "m := {\n\ta: 1\n\t// last\n}\n",
			want:  "m := {\n\ta: 1\n\t// last\n}\n",
		},
		{
			name:  "multiple comments",
			input: "m := {\n\t// c1\n\ta: 1,\n\t// c2\n\tb: 2\n\t// c3\n}\n",
			want:  "m := {\n\t// c1\n\ta: 1,\n\t// c2\n\tb: 2\n\t// c3\n}\n",
		},
	}
	// Comments inside maps force expansion regardless of length.
	runCases(t, fmtDefault, cases)
}

// TestFormat_MapInline verifies that short maps are collapsed to one line and
// long maps are kept expanded, and that maps with comments always expand.
func TestFormat_MapInline(t *testing.T) {
	inlineCases := []struct{ name, input, want string }{
		{
			name:  "empty map stays inline",
			input: "m := {}\n",
			want:  "m := {}\n",
		},
		{
			name:  "short single-element map inlined",
			input: "m := {\n\ta: 1\n}\n",
			want:  "m := {a: 1}\n",
		},
		{
			name:  "short multi-element map inlined",
			input: "m := {\n\ta: 1,\n\tb: 2\n}\n",
			want:  "m := {a: 1, b: 2}\n",
		},
		{
			name:  "map with comment is NOT inlined",
			input: "m := {\n\t// note\n\ta: 1\n}\n",
			want:  "m := {\n\t// note\n\ta: 1\n}\n",
		},
	}
	runCases(t, fmtDefault, inlineCases)

	// Always-expand mode.
	expandCases := []struct{ name, input, want string }{
		{
			name:  "short map expanded when maxLineLen=0",
			input: "m := {\n\ta: 1\n}\n",
			want:  "m := {\n\ta: 1\n}\n",
		},
	}
	runCases(t, fmtExpand, expandCases)

	// Map too long to inline stays expanded.
	t.Run("long map not inlined", func(t *testing.T) {
		// Build a map that is definitely > 100 chars when rendered inline.
		input := "m := {\n\tuuid: \"site-AAAAAAAAAA\",\n\tcity: 10,\n\thouseNumber: 5,\n\tstreetCode: \"STREET-001\",\n\tzipCode: \"12345\",\n\tcountry: \"Somewhere\"\n}\n"
		got, err := fmtDefault(input)
		if err != nil {
			t.Fatalf("Format error: %v", err)
		}
		// Multi-line output has more than one newline.
		newlines := 0
		for _, c := range got {
			if c == '\n' {
				newlines++
			}
		}
		if newlines <= 1 {
			t.Errorf("expected multi-line output (>1 newline), got:\n%s", got)
		}
	})
}

// TestFormat_CallExprMultiLine verifies that function calls whose arguments
// span multiple source lines are kept multi-line (one arg per line).
func TestFormat_CallExprMultiLine(t *testing.T) {
	cases := []struct{ name, input, want string }{
		{
			name:  "single-line call stays inline",
			input: "foo(a, b, c)\n",
			want:  "foo(a, b, c)\n",
		},
		{
			name: "multi-line call: one arg per line",
			input: "foo(\n\ta,\n\tb,\n\tc\n)\n",
			want:  "foo(\n\ta,\n\tb,\n\tc\n)\n",
		},
		{
			name: "nested multi-line calls",
			input: "outer(\n\tinner(\n\t\tx,\n\t\ty\n\t),\n\tz\n)\n",
			want:  "outer(\n\tinner(\n\t\tx,\n\t\ty\n\t),\n\tz\n)\n",
		},
		{
			name: "comment between args preserved",
			input: "foo(\n\ta,\n\t// note\n\tb\n)\n",
			want:  "foo(\n\ta,\n\t// note\n\tb\n)\n",
		},
	}
	runCases(t, fmtDefault, cases)
}

// TestFormat_ArrayLit verifies that array literals are inlined when they fit
// within the configured line length and expanded otherwise — mirroring the
// map literal behaviour.
func TestFormat_ArrayLit(t *testing.T) {
	inlineCases := []struct{ name, input, want string }{
		{
			name:  "empty array stays inline",
			input: "a := []\n",
			want:  "a := []\n",
		},
		{
			name:  "single-line array stays inline",
			input: "a := [1, 2, 3]\n",
			want:  "a := [1, 2, 3]\n",
		},
		{
			name:  "short multi-line array collapsed to one line",
			input: "a := [\n\t1,\n\t2,\n\t3\n]\n",
			want:  "a := [1, 2, 3]\n",
		},
		{
			name:  "string elements inlined when short",
			input: "a := [\n\t\"site-A\",\n\t\"site-B\"\n]\n",
			want:  "a := [\"site-A\", \"site-B\"]\n",
		},
		{
			name:  "array with comment is NOT inlined",
			input: "a := [\n\t// note\n\t1,\n\t2\n]\n",
			want:  "a := [\n\t// note\n\t1,\n\t2\n]\n",
		},
	}
	runCases(t, fmtDefault, inlineCases)

	// Always-expand mode.
	expandCases := []struct{ name, input, want string }{
		{
			name:  "short array expanded when maxLineLen=0",
			input: "a := [\n\t1,\n\t2\n]\n",
			want:  "a := [\n\t1,\n\t2\n]\n",
		},
	}
	runCases(t, fmtExpand, expandCases)

	// Array too long to inline stays expanded.
	t.Run("long array not inlined", func(t *testing.T) {
		input := "a := [\n\t\"very-long-string-number-one\",\n\t\"very-long-string-number-two\",\n\t\"very-long-string-number-three\",\n\t\"very-long-string-number-four\"\n]\n"
		got, err := fmtDefault(input)
		if err != nil {
			t.Fatalf("Format error: %v", err)
		}
		newlines := 0
		for _, c := range got {
			if c == '\n' {
				newlines++
			}
		}
		if newlines <= 1 {
			t.Errorf("expected multi-line output (>1 newline), got:\n%s", got)
		}
	})
}

// TestFormat_Idempotent verifies that formatting twice produces the same result.
func TestFormat_Idempotent(t *testing.T) {
	inputs := []string{
		"x := 1\ny := 2\n",
		"m := {a: 1, b: 2}\n",
		"m := {\n\t// comment\n\ta: 1\n}\n",
		"foo(\n\ta,\n\tb\n)\n",
		"x := 0.0\ny := 100.0\n",
		`m := {"site-A": 1.0}` + "\n",
		"a := [1, 2, 3]\n",
		"a := [\n\t1,\n\t2\n]\n",
	}
	for _, src := range inputs {
		t.Run(src, func(t *testing.T) {
			first, err := fmtDefault(src)
			if err != nil {
				t.Fatalf("first Format error: %v", err)
			}
			second, err := fmtDefault(string(first))
			if err != nil {
				t.Fatalf("second Format error: %v", err)
			}
			if string(first) != string(second) {
				t.Errorf("not idempotent:\nfirst:\n%s\nsecond:\n%s", first, second)
			}
		})
	}
}
