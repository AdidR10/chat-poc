# Chat POC - OpenCode Architecture Demo

A streaming chat application demonstrating OpenCode's real-time TUI architecture using modern tools:
- **Backend**: Bun + TypeScript with streaming HTTP responses  
- **Frontend**: Go TUI powered by Bubble Tea and Lipgloss
- **Communication**: Character-by-character streaming similar to AI chat interfaces

## 🏗️ Architecture Overview

This project replicates OpenCode's client-server streaming pattern:

```
┌─────────────────┐    HTTP POST /chat    ┌─────────────────┐
│   Go Frontend   │◄─────────────────────►│ TypeScript      │
│                 │                        │ Backend         │
│ • Bubble Tea    │    Streaming chars     │                 │
│ • Lipgloss      │    ←←←←←←←←←←←←←←←←    │ • Bun Server    │
│ • Text Input    │                        │ • Port 3000     │
│ • Chat History  │                        │ • /chat API     │
└─────────────────┘                        └─────────────────┘
```

## ✨ Key Features

- **🔄 Real-time Streaming**: Watch responses appear character-by-character like ChatGPT
- **🎨 Beautiful Terminal UI**: Professional styling with borders, colors, and layouts  
- **📝 Chat History**: Persistent conversation with distinct user/bot message styling
- **⚡ Event-Driven**: Bubble Tea's reactive architecture for smooth interactions
- **🛡️ Error Handling**: Graceful connection failures and stream interruptions
- **📱 Responsive Design**: Adapts to terminal window resizing

## 📁 Project Structure

```
chat-poc/
├── backend/                 # Bun TypeScript server
│   ├── package.json        # Dependencies: bun
│   ├── server.ts           # HTTP server with streaming /chat endpoint
│   └── node_modules/       # Auto-generated (ignored by git)
│
├── frontend/               # Go TUI application  
│   ├── go.mod             # Dependencies: bubbletea, lipgloss
│   ├── go.sum             # Dependency checksums
│   ├── main.go            # Application entry point
│   └── chat.go            # Core TUI implementation
│
├── .gitignore             # Excludes node_modules, binaries, logs
└── README.md              # This documentation
```

## 🛠️ Prerequisites

Install these tools before running the project:

| Tool | Version | Install |
|------|---------|---------|
| **Bun** | Latest | [bun.sh](https://bun.sh) - Fast JavaScript runtime |
| **Go** | 1.21+ | [golang.org](https://golang.org) - Systems programming language |

## 🚀 Quick Start

### 1️⃣ Clone & Setup
```bash
git clone <your-repo>
cd chat-poc
```

### 2️⃣ Backend Setup
```bash
cd backend
bun install              # Install dependencies
bun run dev             # Start development server
```

✅ **Success output:**
```
🚀 Chat server running on http://localhost:3000
📡 Endpoints available:
   GET  /health - Health check
   POST /chat   - Send chat message (streams response)
```

### 3️⃣ Frontend Setup (new terminal)
```bash
cd frontend
go mod tidy             # Download Go dependencies  
go run .                # Launch TUI application
```

## 🎮 How to Use

1. **💬 Type your message** in the blue input box at the bottom
2. **⏎ Press Enter** to send (input clears automatically)
3. **👀 Watch magic happen**: Response streams in character-by-character
4. **📜 Scroll through history**: See previous messages with color-coded borders
5. **❌ Press Ctrl+C** to quit gracefully

### UI Layout
```
💬 Chat POC - OpenCode Architecture Demo

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

Press Enter to send • Ctrl+C to quit • Characters stream in real-time
```

## ⚙️ Technical Deep Dive

### Backend Architecture (server.ts)
```typescript
// Core streaming implementation
const stream = new ReadableStream({
  start(controller) {
    const response = generateResponse(message);
    let index = 0;
    
    const interval = setInterval(() => {
      if (index < response.length) {
        controller.enqueue(response[index]);  // Send one character
        index++;
      } else {
        controller.close();                   // End stream
        clearInterval(interval);
      }
    }, 50);  // 50ms delay between characters
  }
});
```

### Frontend Architecture (chat.go)
```go
// Bubble Tea message types for streaming
type (
    streamCharMsg string      // Single character received
    streamEndMsg  struct{}    // Stream completed
    streamErrMsg  string      // Error occurred
)

// Update function handles all events
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case streamCharMsg:
        m.currentResponse += string(msg)  // Append character
        return m, nil
    case streamEndMsg:
        m.messages = append(m.messages, "Bot: "+m.currentResponse)
        m.streaming = false
        return m, nil
    }
}
```

## 🔧 Troubleshooting

### 🔴 Backend Issues

**Error: Port 3000 in use**
```bash
lsof -ti:3000 | xargs kill -9    # Kill process on port 3000
bun run dev                      # Restart server
```

**Error: Bun command not found**
```bash
curl -fsSL https://bun.sh/install | bash    # Install Bun
source ~/.bashrc                             # Reload shell
```

### 🔴 Frontend Issues

**Error: Go modules not found**
```bash
cd frontend
go mod tidy                      # Re-download dependencies
go clean -modcache              # Clear module cache if needed
```

**Error: Failed to connect to server**
```bash
# Check if backend is running
curl http://localhost:3000/health

# Should return: {"status": "ok", "timestamp": "..."}
```

### 🔴 Common Issues

| Problem | Solution |
|---------|----------|
| No streaming effect | Ensure both frontend and backend are running |
| UI looks broken | Resize terminal window (minimum 80x24) |
| Characters appear too fast | Increase delay in `server.ts` (line with `50ms`) |
| Terminal colors missing | Use modern terminal (iTerm2, Windows Terminal, etc.) |

## 🎯 Learning Outcomes

After exploring this codebase, you'll understand:

### 🧠 Core Concepts
1. **Streaming HTTP Responses**: How to send data progressively without WebSockets
2. **Event-Driven TUI**: Bubble Tea's Update/View architecture pattern
3. **Terminal Styling**: Professional CLI interfaces with Lipgloss
4. **Client-Server Communication**: REST API design for real-time applications
5. **Concurrent Programming**: Goroutines for non-blocking HTTP streaming

### 🔍 OpenCode Connections
- **TUI Architecture**: Similar to OpenCode's `packages/tui/` structure
- **Streaming Responses**: How AI providers stream tokens to the interface
- **State Management**: Event-driven updates for chat conversations
- **Error Handling**: Graceful degradation when connections fail

## 🚀 Next Steps

### 🔬 Dive Deeper into OpenCode
1. **Explore `packages/tui/`**: Compare TUI implementations
2. **Study AI Provider Integration**: How OpenCode connects to GPT/Claude/etc.
3. **Examine Tool System**: How OpenCode passes tools to AI models
4. **Analyze Session Management**: Conversation persistence and context

### 🛠️ Extend This Project
1. **Add Authentication**: User login/logout system
2. **Multiple Conversations**: Chat history with different sessions
3. **File Upload**: Send files to the chat bot
4. **Custom Themes**: Different color schemes and layouts
5. **WebSocket Upgrade**: Replace HTTP streaming with WebSockets

## 📚 Resources

- [Bubble Tea Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
- [Lipgloss Examples](https://github.com/charmbracelet/lipgloss/tree/master/examples)  
- [Bun Documentation](https://bun.sh/docs)
- [OpenCode Repository](https://github.com/opencodeai/opencode)

---

**🎯 Built for learning OpenCode's architecture** • **⚡ Real streaming responses** • **🎨 Beautiful terminal UI** • **🔄 Event-driven design**

*Happy coding! 🚀*