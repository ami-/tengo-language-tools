package lsp

import "encoding/json"

// Base wire types

type RequestMessage struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"` // nil = notification
	Method  string           `json:"method"`
	Params  json.RawMessage  `json:"params,omitempty"`
}

type ResponseMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *ResponseError  `json:"error,omitempty"`
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type NotificationMessage struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

// LSP error codes
const (
	CodeParseError           = -32700
	CodeInvalidRequest       = -32600
	CodeMethodNotFound       = -32601
	CodeServerNotInitialized = -32002
)

// initialize

type InitializeParams struct {
	ProcessID             *int   `json:"processId"`
	RootURI               string `json:"rootUri"`
	RootPath              string `json:"rootPath"` // legacy fallback
	InitializationOptions *struct {
		MaxLineLen int `json:"maxLineLen"`
	} `json:"initializationOptions,omitempty"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   ServerInfo         `json:"serverInfo"`
}

type ServerCapabilities struct {
	TextDocumentSync           int                `json:"textDocumentSync"` // 1 = Full
	HoverProvider              bool               `json:"hoverProvider"`
	ReferencesProvider         bool               `json:"referencesProvider"`
	DocumentSymbolProvider     bool               `json:"documentSymbolProvider"`
	DocumentFormattingProvider bool               `json:"documentFormattingProvider"`
	DefinitionProvider         bool               `json:"definitionProvider"`
	CompletionProvider         *CompletionOptions `json:"completionProvider,omitempty"`
	RenameProvider             bool               `json:"renameProvider"`
}

type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters"`
}

// formatting

type DocumentFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Options      FormattingOptions      `json:"options"`
}

type FormattingOptions struct {
	TabSize      int  `json:"tabSize"`
	InsertSpaces bool `json:"insertSpaces"`
}

type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// text sync

type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type TextDocumentItem struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
	Text    string `json:"text"`
}

type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

type TextDocumentContentChangeEvent struct {
	Text string `json:"text"` // Full sync: always complete text
}

type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

// hover

type HoverParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

type HoverResult struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

type MarkupContent struct {
	Kind  string `json:"kind"` // "markdown" or "plaintext"
	Value string `json:"value"`
}

// references

type ReferenceParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	Context      ReferenceContext       `json:"context"`
}

type ReferenceContext struct {
	IncludeDeclaration bool `json:"includeDeclaration"`
}

type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// document symbols

type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type DocumentSymbol struct {
	Name           string           `json:"name"`
	Kind           SymbolKind       `json:"kind"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

type SymbolKind int

const (
	SymbolKindFunction SymbolKind = 12
	SymbolKindVariable SymbolKind = 13
)

// completion

type CompletionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

type CompletionItem struct {
	Label         string `json:"label"`
	Kind          int    `json:"kind"` // 3=Function 6=Variable 9=Module 14=Keyword
	Detail        string `json:"detail,omitempty"`
	Documentation string `json:"documentation,omitempty"`
}

// rename

type RenameParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	NewName      string                 `json:"newName"`
}

type PrepareRenameParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

type WorkspaceEdit struct {
	Changes map[string][]TextEdit `json:"changes"`
}

// diagnostics

type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Version     *int         `json:"version,omitempty"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"` // 1=Error 2=Warning 3=Information 4=Hint
	Message  string `json:"message"`
	Source   string `json:"source"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Position struct {
	Line      int `json:"line"`      // 0-based
	Character int `json:"character"` // 0-based
}
