// A small sample MCP server built with github.com/modelcontextprotocol/go-sdk/mcp.
//
// It exposes one of each core primitive:
//   - tools:     "calculate" and "current_time"
//   - resource:  "info://server" (static text)
//   - prompt:    "code_review"
//
// Run over stdio (default, for Claude Desktop / Claude Code / Inspector):
//
//	go run .
//
// Run over streamable HTTP:
//
//	go run . -http :8080
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ---- Tool: calculate ------------------------------------------------------

// The SDK derives the tool's input JSON schema from this struct, and validates
// incoming arguments against it before calling the handler.
type CalcInput struct {
	Op string  `json:"op" jsonschema:"one of: add, sub, mul, div"`
	A  float64 `json:"a" jsonschema:"first operand"`
	B  float64 `json:"b" jsonschema:"second operand"`
}

// The output schema is derived from this struct and returned as structured content.
type CalcOutput struct {
	Result float64 `json:"result" jsonschema:"result of the operation"`
}

func calculate(ctx context.Context, req *mcp.CallToolRequest, in CalcInput) (*mcp.CallToolResult, CalcOutput, error) {
	var r float64
	switch in.Op {
	case "add":
		r = in.A + in.B
	case "sub":
		r = in.A - in.B
	case "mul":
		r = in.A * in.B
	case "div":
		if in.B == 0 {
			// Returning an error makes the SDK send a tool result with IsError=true,
			// so the LLM can see the message and recover.
			return nil, CalcOutput{}, fmt.Errorf("division by zero")
		}
		r = in.A / in.B
	default:
		return nil, CalcOutput{}, fmt.Errorf("unknown op %q (want add, sub, mul, div)", in.Op)
	}
	// Returning a nil *CallToolResult lets the SDK build the result (text + structured
	// content) from the Out value.
	return nil, CalcOutput{Result: r}, nil
}

// ---- Tool: current_time ---------------------------------------------------

type TimeInput struct {
	Timezone string `json:"timezone,omitempty" jsonschema:"IANA timezone, e.g. Asia/Singapore (default UTC)"`
}

func currentTime(ctx context.Context, req *mcp.CallToolRequest, in TimeInput) (*mcp.CallToolResult, any, error) {
	name := in.Timezone
	if name == "" {
		name = "UTC"
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid timezone %q: %w", name, err)
	}
	// Building the result by hand, with plain text content.
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: time.Now().In(loc).Format(time.RFC1123)},
		},
	}, nil, nil
}

// ---- Resource -------------------------------------------------------------

const infoURI = "info://server"

func readInfo(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	if req.Params.URI != infoURI {
		return nil, mcp.ResourceNotFoundError(req.Params.URI)
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      infoURI,
			MIMEType: "text/plain",
			Text:     "sample-mcp-server: a demo built with the official Go SDK.",
		}},
	}, nil
}

// ---- Prompt ---------------------------------------------------------------

func codeReview(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	code := req.Params.Arguments["code"]
	return &mcp.GetPromptResult{
		Description: "Ask for a code review",
		Messages: []*mcp.PromptMessage{{
			Role: "user",
			Content: &mcp.TextContent{
				Text: "Please review this code for bugs, clarity and style:\n\n" + code,
			},
		}},
	}, nil
}

// ---- Wiring ---------------------------------------------------------------

func newServer() *mcp.Server {
	s := mcp.NewServer(
		&mcp.Implementation{Name: "sample-mcp-server", Version: "v0.1.0"},
		&mcp.ServerOptions{Instructions: "Use calculate for arithmetic and current_time for the time."},
	)

	mcp.AddTool(s, &mcp.Tool{Name: "calculate", Description: "Basic arithmetic on two numbers"}, calculate)
	mcp.AddTool(s, &mcp.Tool{Name: "current_time", Description: "Current time in a timezone"}, currentTime)

	s.AddResource(&mcp.Resource{
		Name:        "server-info",
		URI:         infoURI,
		MIMEType:    "text/plain",
		Description: "About this server",
	}, readInfo)

	s.AddPrompt(&mcp.Prompt{
		Name:        "code_review",
		Description: "Review a snippet of code",
		Arguments:   []*mcp.PromptArgument{{Name: "code", Description: "Code to review", Required: true}},
	}, codeReview)

	return s
}

func main() {
	httpAddr := flag.String("http", "", "listen address for streamable HTTP (e.g. :8080); empty = stdio")
	flag.Parse()

	if *httpAddr != "" {
		handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return newServer() }, nil)
		log.Printf("MCP server listening on %s", *httpAddr)
		log.Fatal(http.ListenAndServe(*httpAddr, handler))
	}

	// stdio: stdout is the protocol channel, so log to stderr only.
	log.SetOutput(os.Stderr)
	if err := newServer().Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
