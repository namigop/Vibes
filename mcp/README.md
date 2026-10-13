# Sample MCP server (Go)

```bash
go mod init example.com/sample-mcp-server   # once
go get github.com/modelcontextprotocol/go-sdk@latest
go mod tidy

go run .                 # stdio
go run . -http :8080     # streamable HTTP at http://localhost:8080
```

## Try it with the MCP Inspector
```bash
npx @modelcontextprotocol/inspector go run .
```

## Use with Claude Code
```bash
go build -o sample-mcp-server .
claude mcp add sample -- /absolute/path/to/sample-mcp-server
```

## Use with Claude Desktop (claude_desktop_config.json)
```json
{ "mcpServers": { "sample": { "command": "/absolute/path/to/sample-mcp-server" } } }
```
