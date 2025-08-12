## Chat POC – OpenCode Architecture Demo

A streaming chat application demonstrating OpenCode's real-time TUI architecture using modern tools:
- **Backend**: Bun + TypeScript with streaming HTTP responses and bash tool execution
- **Frontend**: Go TUI powered by Bubble Tea and Lipgloss
- **Communication**: Character-by-character streaming similar to AI chat interfaces
- **Tool Calls**: Execute system commands directly from chat (like OpenCode's tool system)

## Chat POC Demo
![Terminal Demo](resources/demo.gif)

## 🏗️ Architecture Overview

This project replicates OpenCode's client-server streaming pattern with tool execution:

```text
┌─────────────────┐    HTTP POST /chat    ┌─────────────────┐
│   Go Frontend   │◄─────────────────────►│   TypeScript     │
│                 │                        │     Backend      │
│ • Bubble Tea    │    Streaming chars     │ • Bun Server     │
│ • Lipgloss      │    ←←←←←←←←←←←←←←←←    │ • Port 3000      │
│ • Text Input    │                        │ • /chat API      │
│ • Chat History  │                        │                  │
└─────────────────┘                        └─────────────────┘
```

## ✨ Key Features

- **🔄 Real-time Streaming**: Watch responses appear character-by-character like ChatGPT
- **🎨 Beautiful Terminal UI**: Professional styling with borders, colors, and layouts
- **📝 Chat History**: Persistent conversation with distinct user/bot message styling
- **🔧 Tool Execution**: Run system commands from chat (`ls`, `pwd`, `df`, `uname`)
- **⚡ Event-Driven**: Bubble Tea's reactive architecture for smooth interactions
- **🛡️ Error Handling**: Graceful connection failures and stream interruptions
- **🔒 Safe Execution**: Whitelisted commands only for security
- **📱 Responsive Design**: Adapts to terminal window resizing

## 📁 Project Structure

```text
chat-poc/
├── backend/                  # Bun TypeScript server
│   ├── package.json          # Dependencies: bun
│   └── server.ts             # HTTP server with streaming /chat endpoint + tool execution
│
├── frontend/                 # Go TUI application
│   ├── go.mod                # Dependencies: bubbletea, lipgloss
│   ├── go.sum                # Dependency checksums
│   ├── main.go               # Application entry point
│   └── chat.go               # Core TUI implementation with tool display
│
├── resources/                # Additional assets
│   ├── demo.cast
│   ├── demo.gif  
└── README.md                 # This documentation
```

## 🛠️ Prerequisites

Install these tools before running the project:

| Tool | Version | Install |
|------|---------|---------|
| **Bun** | Latest | [`bun.sh`](https://bun.sh) – Fast JavaScript runtime |
| **Go** | 1.21+ | [`golang.org`](https://golang.org) – Systems programming language |

## 🚀 Quick Start

### 1. Clone the repo

```bash
git clone <your-repo>
cd chat-poc
```

### 2. Backend setup

```bash
cd backend
bun install          # Install dependencies
bun run dev          # Start development server
```

Expected output (abridged):

```text
🚀 Chat server running on http://localhost:3000
📡 Endpoints:
  GET  /health
  POST /chat
```

### 3. Frontend setup (new terminal)

```bash
cd frontend
go mod tidy          # Download Go dependencies
go run .             # Launch TUI application
```

## 🎮 How to Use

- **Basic chat**
  - Type your message in the purple input box at the bottom
  - Press Enter to send (input clears automatically)
  - Responses stream in character-by-character
  - Scroll through history with color-coded borders
  - Press Ctrl+C to quit gracefully

- **Tool commands** (try these):

| Command | Triggers | What it does |
|---|---|---|
| List Files | "list files", "show files", "ls" | Shows directory contents with `ls -la` |
| Current Directory | "current directory", "pwd" | Shows working directory with `pwd` |
| Disk Usage | "disk usage", "df" | Shows disk space with `df -h` |
| System Info | "system info", "uname" | Shows system details with `uname -a` |

### Example chat session

You: What files are in this directory?

Bot: I'll execute the `ls` command for you...

```text
🔧 Executing: ls
total 16
drwxr-xr-x 4 user staff 128 Nov 15 10:30 .
drwxr-xr-x 5 user staff 160 Nov 15 10:25 ..
-rw-r--r-- 1 user staff 245 Nov 15 10:30 README.md
drwxr-xr-x 3 user staff 96 Nov 15 10:28 backend
drwxr-xr-x 3 user staff 96 Nov 15 10:28 frontend
```

Based on the output, I can see 5 items in the current directory.

### UI layout with tool calls

```text
💬 Chat POC – OpenCode Architecture Demo

╭─────────────────────────────────────────╮
│ You: Hello there!                       │  ← Green border (your messages)
╰─────────────────────────────────────────╯

╭─────────────────────────────────────────╮
│ Bot: Hi! How can I help you today?      │  ← Red border (bot messages)
╰─────────────────────────────────────────╯

╭═════════════════════════════════════════╮
│ Bot: I'm currently typing...▎           │  ← Yellow double border (streaming)
╰═════════════════════════════════════════╯

╭─────────────────────────────────────────╮
│ Type your message here...               │  ← Purple border (input)
╰─────────────────────────────────────────╯

╭─────────────────────────────────────────╮
│ You: Show me the files here             │
╰─────────────────────────────────────────╯

╭─────────────────────────────────────────╮
│ Bot: I'll execute the ls command for    │
│ you...                                  │
│                                         │
│ 🔧 Executing: ls                        │
│ total 16                                │
│ drwxr-xr-x  4 user  staff  128 ...      │
│ -rw-r--r--  1 user  staff  245 ...      │
│                                         │
│ Based on the output, I can see 5 items  │
│ in the current directory.               │
╰─────────────────────────────────────────╯

Press Enter to send • Ctrl+C to quit • Characters stream in real-time
```

## ⚙️ Technical Deep Dive

### Backend architecture (`backend/server.ts`)

#### Tool detection and execution

```ts
// Pattern matching for tool triggers
function detectToolCall(message: string): ToolCall | null {
  const patterns = [
    { regex: /list files|show files|ls/i, command: "ls", args: ["-la"] },
    { regex: /current directory|pwd/i, command: "pwd", args: [] },
    { regex: /disk usage|df/i, command: "df", args: ["-h"] },
    { regex: /system info|uname/i, command: "uname", args: ["-a"] },
  ];

  for (const pattern of patterns) {
    if (pattern.regex.test(message)) {
      return { command: pattern.command, args: pattern.args };
    }
  }
  return null;
}

// Safe command execution with whitelisting
async function executeTool(tool: ToolCall): Promise<string> {
  const allowedCommands = ["ls", "pwd", "df", "uname"];

  if (!allowedCommands.includes(tool.command)) {
    return `Command '${tool.command}' is not allowed`;
  }

  // Execute using Bun's $ shell operator
  const result = await $`${tool.command} ${tool.args}`.text();
  return result.trim();
}
```

#### Streaming with tool markers

```ts
// Tool execution flow in streaming response
if (tool) {
  // 1. Stream pre-message
  const preMessage = `I'll execute the ${tool.command} command for you...\n\n`;
  for (const char of preMessage) {
    controller.enqueue(encoder.encode(char));
    await new Promise((resolve) => setTimeout(resolve, 20));
  }

  // 2. Send tool start marker
  controller.enqueue(encoder.encode(`[TOOL_START:${tool.command}]`));

  // 3. Execute tool
  const output = await executeTool(tool);

  // 4. Send tool output
  controller.enqueue(encoder.encode(`[TOOL_OUTPUT:${output}]`));

  // 5. Send tool end marker
  controller.enqueue(encoder.encode("[TOOL_END]"));

  // 6. Stream analysis
  const analysis = `\n\nBased on the output, ${analyzeToolOutput(tool.command, output)}`;
  for (const char of analysis) {
    controller.enqueue(encoder.encode(char));
    await new Promise((resolve) => setTimeout(resolve, 20));
  }
}
```

### Frontend architecture (`frontend/chat.go`)

#### Tool message types

```go
// New message types for tool calls
type (
    streamCharMsg     string      // Regular character streaming
    streamEndMsg      struct{}    // Stream completed
    streamErrMsg      string      // Stream error
    toolCallStartMsg  string      // Tool command being executed
    toolCallOutputMsg string      // Tool output
    toolCallEndMsg    struct{}    // Tool execution complete
)
```

#### Tool stream parsing

```go
// In sendMessage goroutine - parse tool markers from stream
bufferStr := buffer.String()

// Tool start marker: [TOOL_START:ls]
if strings.HasPrefix(bufferStr, "[TOOL_START:") && strings.Contains(bufferStr, "]") {
    endIdx := strings.Index(bufferStr, "]")
    toolCmd := bufferStr[12:endIdx] // Extract command
    globalProgram.Send(toolCallStartMsg(toolCmd))
    buffer.Reset()
    continue
}

// Tool output marker: [TOOL_OUTPUT:file1.txt\nfile2.go]
if strings.HasPrefix(bufferStr, "[TOOL_OUTPUT:") && strings.Contains(bufferStr, "]") {
    endIdx := strings.Index(bufferStr, "]")
    output := bufferStr[13:endIdx] // Extract output
    globalProgram.Send(toolCallOutputMsg(output))
    buffer.Reset()
    continue
}
```

#### Tool UI rendering

```go
// Handle tool messages in Update function
case toolCallStartMsg:
    m.currentResponse += fmt.Sprintf("\n\n🔧 Executing: %s\n", string(msg))
    return m, nil

case toolCallOutputMsg:
    // Format tool output in code block
    m.currentResponse += fmt.Sprintf("```\n%s\n```\n", string(msg))
    return m, nil

case toolCallEndMsg:
    m.currentResponse += "\n"
    return m, nil
```

## 🔧 Troubleshooting

### Backend issues

- **Error: Port 3000 in use**

```bash
lsof -ti:3000 | xargs kill -9    # Kill process on port 3000
bun run dev                      # Restart server
```

- **Error: Bun command not found**

```bash
curl -fsSL https://bun.sh/install | bash    # Install Bun
source ~/.bashrc                            # Reload shell
```

- **Error: Tool execution failed**

```bash
# Check if commands are available
ls --version
pwd --version
df --version
uname --version
```

### Frontend issues

- **Error: Go modules not found**

```bash
cd frontend
go mod tidy                # Re-download dependencies
go clean -modcache         # Clear module cache if needed
```

- **Error: Failed to connect to server**

```bash
# Check if backend is running
curl http://localhost:3000/health
# Should return: Server is running!
```

- **Error: Tool calls not working**

```bash
# Test tool detection manually
curl -X POST http://localhost:3000/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "list files"}'
```

### Common issues

| Problem | Solution |
|---|---|
| No tool execution | Try exact phrases: "list files", "current directory" |
| Tool output garbled | Ensure terminal supports UTF-8 encoding |
| Commands not found | Check if `ls`, `pwd`, `df`, `uname` are in PATH |
| Permission denied | Some commands may require different permissions |
| Streaming stops during tools | Backend might have crashed – check console |

## 🎯 Learning Outcomes

After exploring this codebase, you'll understand:

- **Streaming HTTP Responses**: How to send data progressively without WebSockets
- **Event-Driven TUI**: Bubble Tea's Update/View architecture pattern
- **Terminal Styling**: Professional CLI interfaces with Lipgloss
- **Tool Execution**: Safe command execution with whitelisting
- **Protocol Design**: Custom markers for tool communication
- **Concurrent Programming**: Goroutines for non-blocking HTTP streaming

## 🔍 OpenCode Connections

- **TUI Architecture**: Similar to OpenCode's `packages/tui/` structure
- **Streaming Responses**: How AI providers stream tokens to the interface
- **Tool System**: How OpenCode executes tools and displays results
- **State Management**: Event-driven updates for chat conversations
- **Security**: Command whitelisting and safe execution patterns