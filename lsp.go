package main

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

var assemblerInstructions = []string{
    "ADC", "SBC",
    "DEC", "DEX", "DEY",
    "INC", "INX", "INY",
    "LDA", "LDX", "LDY",
    "STA", "STX", "STY",
    "TAX", "TAY", "TXS", "TXA", "TYA", "TSX",
    "PHA", "PHP",
    "PLA", "PLP",
    "CMP", "CPX", "CPY",
}

type server struct {
    stdout *bufio.Writer
}

func main() {
    // Set up logging to a file
    f, err := os.OpenFile("/tmp/lsp.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatalf("Failed to open log file: %v", err)
    }
    defer f.Close()
    log.SetOutput(f)

    log.Println("Starting LSP server...")

    // Use a buffered writer to ensure responses are flushed
    stdout := bufio.NewWriter(os.Stdout)
    stream := jsonrpc2.NewBufferedStream(stdrwc{}, jsonrpc2.VSCodeObjectCodec{})
    conn := jsonrpc2.NewConn(context.Background(), stream, server{stdout: stdout})
    defer conn.Close()

    log.Println("LSP server started.")
    <-conn.DisconnectNotify()
}

func (s server) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
    log.Printf("Received request: %+v", req)

    switch req.Method {
    case "initialize":
        var params lsp.InitializeParams
        if err := json.Unmarshal(*req.Params, &params); err != nil {
            log.Printf("Failed to unmarshal initialize params: %v", err)
            conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
                Code:    jsonrpc2.CodeInvalidParams,
                Message: "Invalid initialize parameters",
            })
            return
        }
        result := lsp.InitializeResult{
            Capabilities: lsp.ServerCapabilities{
                TextDocumentSync: &lsp.TextDocumentSyncOptionsOrKind{
                    Options: &lsp.TextDocumentSyncOptions{
                        OpenClose: true,
                        Change:    lsp.TDSKFull,
                    },
                },
                CompletionProvider: &lsp.CompletionOptions{
                    ResolveProvider: false,
                    TriggerCharacters: []string{" "}, // Trigger on space
                },
            },
        }
        log.Printf("Sending initialize result: %+v", result)
        if err := conn.Reply(ctx, req.ID, result); err != nil {
            log.Printf("Failed to send initialize result: %v", err)
        }

        s.stdout.Flush() // Ensure the output is flushed
    case "textDocument/didOpen", "textDocument/didChange":
        // Handle document opening and changes if needed
        log.Printf("Handling %s request", req.Method)
        var params lsp.DidOpenTextDocumentParams
        if err := json.Unmarshal(*req.Params, &params); err != nil {
            log.Printf("Failed to unmarshal didOpen params: %v", err)
            return
        }
        log.Printf("Document opened: %s", params.TextDocument.URI)
    case "textDocument/completion":
        var params lsp.CompletionParams
        if err := json.Unmarshal(*req.Params, &params); err != nil {
            log.Printf("Failed to unmarshal completion params: %v", err)
            conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
                Code:    jsonrpc2.CodeInvalidParams,
                Message: "Invalid completion parameters",
            })
            return
        }
        items := []lsp.CompletionItem{}
        for _, instr := range assemblerInstructions {
            items = append(items, lsp.CompletionItem{
                Label: instr,
                Kind:  lsp.CIKOperator,
            })
        }
        log.Printf("Sending completion result: %+v", items)
        if err := conn.Reply(ctx, req.ID, lsp.CompletionList{
            IsIncomplete: false,
            Items:        items,
        }); err != nil {
            log.Printf("Failed to send completion result: %v", err)
        }
        s.stdout.Flush() // Ensure the output is flushed
    case "shutdown":
        log.Println("Received shutdown request")
        if err := conn.Reply(ctx, req.ID, nil); err != nil {
            log.Printf("Failed to send shutdown response: %v", err)
        }
        s.stdout.Flush() // Ensure the output is flushed
    case "exit":
        log.Println("Received exit request")
        os.Exit(0)
    default:
        log.Printf("Unknown method: %s", req.Method)
        if err := conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
            Code:    jsonrpc2.CodeMethodNotFound,
            Message: "Method not found",
        }); err != nil {
            log.Printf("Failed to send error response: %v", err)
        }
        s.stdout.Flush() // Ensure the output is flushed
    }
}

type stdrwc struct{}

func (stdrwc) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

func (stdrwc) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

func (stdrwc) Close() error {
	if err := os.Stdin.Close(); err != nil {
		return err
	}
	return os.Stdout.Close()
}
