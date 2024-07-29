package main

import (
	"context"
	"encoding/json"
        "fmt"
	"log"
	"os"
        "sort"
        "sync"
        "strings"
        "strconv"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

var oldAssemblerInstructions = []string{
    "ADC", "SBC",
    "DEC", "DEX", "DEY",
    "INC", "INX", "INY",
    "LDA", "LDX", "LDY",
    "STA", "STX", "STY",
    "TAX", "TAY", "TXS", "TXA", "TYA", "TSX",
    "PHA", "PHP",
    "PLA", "PLP",
    "CMP", "CPX", "CPY",
    "AND", "ORA", "EOR", "BIT",
}

var assemblerInstructions []lsp.CompletionItem

type server struct {
    mu        sync.Mutex
    cancelMap map[jsonrpc2.ID]context.CancelFunc
    documents map[lsp.DocumentURI]string
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
    stream := jsonrpc2.NewBufferedStream(stdrwc{}, jsonrpc2.VSCodeObjectCodec{})
    conn := jsonrpc2.NewConn(context.Background(), stream, &server{
        cancelMap: make(map[jsonrpc2.ID]context.CancelFunc),
        documents: make(map[lsp.DocumentURI]string),
    })
    defer conn.Close()

    log.Println("LSP server started.")
    <-conn.DisconnectNotify()
}

func (s *server) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
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
        err := json.Unmarshal(instructionsJsonString, &assemblerInstructions)
        if err != nil {
		log.Println("error:", err)
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
    case "initialized":
        log.Printf("Handling %s request", req.Method)
    case "textDocument/didOpen":
        var params lsp.DidOpenTextDocumentParams
        if err := json.Unmarshal(*req.Params, &params); err != nil {
            log.Printf("Failed to unmarshal didOpen params: %v", err)
            return
        }
        s.mu.Lock()
        s.documents[params.TextDocument.URI] = params.TextDocument.Text
        s.mu.Unlock()
        log.Printf("Document opened: %s", params.TextDocument.URI)
    case "textDocument/didChange":
        var params lsp.DidChangeTextDocumentParams
        if err := json.Unmarshal(*req.Params, &params); err != nil {
            log.Printf("Failed to unmarshal didChange params: %v", err)
            return
        }
        s.mu.Lock()
        s.documents[params.TextDocument.URI] = params.ContentChanges[len(params.ContentChanges)-1].Text
        s.mu.Unlock()
    case "textDocument/completion":
        ctx, cancel := context.WithCancel(ctx)
        s.mu.Lock()
        s.cancelMap[req.ID] = cancel
        s.mu.Unlock()

        go func() {
            s.handleCompletion(ctx, conn, req)
            s.mu.Lock()
            delete(s.cancelMap, req.ID)
            s.mu.Unlock()
        }()
    case "completionItem/resolve":
        var item lsp.CompletionItem
        if err := json.Unmarshal(*req.Params, &item); err != nil {
            log.Printf("Failed to unmarshal completion item resolve params: %v", err)
            conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
                Code:    jsonrpc2.CodeInvalidParams,
                Message: "Invalid completion item parameters",
            })
            return
        }

        log.Printf("Resolved completion item: %+v", item)
        if err := conn.Reply(ctx, req.ID, item); err != nil {
            log.Printf("Failed to send resolved completion item: %v", err)
        }
    case "$/cancelRequest":
        var params struct {
            ID jsonrpc2.ID `json:"id"`
        }
        if err := json.Unmarshal(*req.Params, &params); err != nil {
            log.Printf("Failed to unmarshal cancel request params: %v", err)
            return
        }
        s.mu.Lock()
        if cancel, ok := s.cancelMap[params.ID]; ok {
            cancel()
            delete(s.cancelMap, params.ID)
        }
        s.mu.Unlock()
    case "shutdown":
        log.Println("Received shutdown request")
        if err := conn.Reply(ctx, req.ID, nil); err != nil {
            log.Printf("Failed to send shutdown response: %v", err)
        }
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
    }
}

func (s *server) handleCompletion(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
    select {
    case <-ctx.Done():
        log.Printf("Request %v cancelled", req.ID)
        return
    default:
    }

    var params lsp.CompletionParams
    if err := json.Unmarshal(*req.Params, &params); err != nil {
        log.Printf("Failed to unmarshal completion params: %v", err)
        conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
            Code:    jsonrpc2.CodeInvalidParams,
            Message: "Invalid completion parameters",
        })
        return
    }

    s.mu.Lock()
    documentText, ok := s.documents[params.TextDocument.URI]
    s.mu.Unlock()

    if !ok {
        log.Printf("Document not found: %s", params.TextDocument.URI)
        conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
            Code:    jsonrpc2.CodeInternalError,
            Message: "Document not found",
        })
        return
    }

    items := []lsp.CompletionItem{}
    userInput := getTextUpToPosition(documentText, params.Position)
    userInput = strings.TrimSpace(userInput)


    if strings.HasPrefix(userInput, "#$") {
        log.Println("PING")
        // Generate hex value suggestions
        for i := 0; i <= 255; i++ {
            hexValue := fmt.Sprintf("#$%02X", i)
            if strings.HasPrefix(hexValue, userInput) {
                items = append(items, lsp.CompletionItem{
                    Label: hexValue,
                    Kind:  lsp.CIKValue,
                })
            }
        }
    } else {
        log.Println("PONG")
        // Check if the last word can be parsed as an integer and convert to hex
        if intValue, err := strconv.Atoi(userInput); err == nil {
            hexValue := fmt.Sprintf("#$%02X", intValue)
            items = append(items, lsp.CompletionItem{
                Label: hexValue,
                Kind:  lsp.CIKValue,
                Detail: fmt.Sprintf("Hex value of %d", intValue),
            })
        }
	for _, instr := range assemblerInstructions {
	    if strings.HasPrefix(instr.Label, userInput) {
		items = append(items, instr)
            }
	}
    }

    sort.Slice(items, func(i, j int) bool {
        return items[i].Label < items[j].Label
    })

    log.Printf("Sending completion result: %+v", items)
    if err := conn.Reply(ctx, req.ID, lsp.CompletionList{
        IsIncomplete: false,
        Items:        items,
    }); err != nil {
        log.Printf("Failed to send completion result: %v", err)
    }
}

func getTextUpToPosition(text string, pos lsp.Position) string {
    lines := strings.Split(text, "\n")
    if int(pos.Line) >= len(lines) {
        return ""
    }
    line := lines[pos.Line]
    if int(pos.Character) > len(line) {
        return line
    }
    return line[:pos.Character]
}

func getLastWord(s string) string {
    words := strings.Fields(s)
    if len(words) == 0 {
        return ""
    }
    return words[len(words)-1]
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
