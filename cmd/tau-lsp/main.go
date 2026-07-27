package main

import (
	"context"
	"io"
	"log"
	"os"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"

	"github.com/NicoNex/tau/internal/parser"
	"github.com/NicoNex/tau/internal/tauerr"
)

// Server represents the LSP server.
type Server struct {
	jsonrpc2.Conn
}

// Initialize the server and set its capabilities.
func (s *Server) Initialize(ctx context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	log.Println("Initializing Tau Language Server...")

	return &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync: &protocol.TextDocumentSyncOptions{
				Change:    protocol.TextDocumentSyncKindIncremental,
				OpenClose: true,
			},
			CompletionProvider: &protocol.CompletionOptions{
				TriggerCharacters: []string{"."},
			},
			SemanticTokensProvider: &protocol.SemanticTokensLegend{
				TokenTypes: []protocol.SemanticTokenTypes{
					protocol.SemanticTokenKeyword,
					protocol.SemanticTokenVariable,
					protocol.SemanticTokenFunction,
					protocol.SemanticTokenString,
					protocol.SemanticTokenNumber,
					protocol.SemanticTokenOperator,
					protocol.SemanticTokenComment,
					protocol.SemanticTokenParameter,
				},
				TokenModifiers: []protocol.SemanticTokenModifiers{
					protocol.SemanticTokenModifierDeclaration,
				},
			},
		},
	}, nil
}

// Handle document open events.
func (s *Server) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	log.Printf("Document opened: %s", params.TextDocument.URI)
	// Perform validation on the opened document
	s.validateTextDocument(ctx, params.TextDocument.URI, params.TextDocument.Text)
	return nil
}

// Handle document change events.
func (s *Server) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	log.Printf("Document changed: %s", params.TextDocument.URI)
	// Perform validation on the changed document
	for _, change := range params.ContentChanges {
		s.validateTextDocument(ctx, params.TextDocument.URI, change.Text)
	}
	return nil
}

// validateTextDocument parses the document and sends diagnostics to the client.
func (s *Server) validateTextDocument(ctx context.Context, uri protocol.DocumentURI, text string) {
	// Call your existing parser (replace with your actual parser function)
	_, errs := parser.Parse("file", text)

	var diagnostics []protocol.Diagnostic
	for _, perr := range errs {
		err, ok := perr.(tauerr.TauErr)
		if !ok {
			continue
		}

		diagnostic := protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      uint32(err.Line - 1), // LSP is zero-based
					Character: uint32(err.Column),
				},
				End: protocol.Position{
					Line:      uint32(err.Line - 1), // LSP is zero-based
					Character: uint32(err.Column),
				},
			},
			Severity: protocol.DiagnosticSeverityError,
			Source:   "Tau Language Server",
			Message:  err.Message,
		}
		diagnostics = append(diagnostics, diagnostic)
	}

	// Publish diagnostics to the client
	params := protocol.PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	}

	err := s.Conn.Notify(
		ctx,
		protocol.MethodTextDocumentPublishDiagnostics,
		params,
	)
	if err != nil {
		log.Println("Failed to publish diagnostics:", err)
	}
}

// Handle completion requests (optional example).
func (s *Server) Completion(ctx context.Context, params *protocol.CompletionParams) (*protocol.CompletionList, error) {
	items := []protocol.CompletionItem{
		// {
		// 	Label:      "if",
		// 	Kind:       protocol.CompletionItemKindKeyword,
		// 	InsertText: protocol.StringPtr("if "),
		// },
		// {
		// 	Label:      "for",
		// 	Kind:       protocol.CompletionItemKindKeyword,
		// 	InsertText: protocol.StringPtr("for "),
		// },
		// Add more completion items as needed
	}

	return &protocol.CompletionList{
		IsIncomplete: false,
		Items:        items,
	}, nil
}

type stdio struct {
	io.ReadCloser
	io.Writer
}

func main() {
	stream := jsonrpc2.NewStream(stdio{
		ReadCloser: os.Stdin,
		Writer:     os.Stdout,
	})

	conn := jsonrpc2.NewConn(stream)
	protocol.ServerDispatcher(conn, nil)
	<-conn.Done()
}
