/**
 * Chat backend with Bus architecture (OpenCode-inspired)
 * Bun runtime
 */

import { $ } from "bun";

// ----------------------
// 🚌 Event Bus
// ----------------------
export type Event<T = any> = {
  type: string;
  data?: T;
  timestamp?: number;
};

type Subscriber = (event: Event) => void;

class EventBus {
  private subscribers: { [type: string]: Subscriber[] } = {};

  subscribe(type: string, handler: Subscriber) {
    if (!this.subscribers[type]) this.subscribers[type] = [];
    this.subscribers[type].push(handler);
    return () => {
      this.subscribers[type] = this.subscribers[type].filter(h => h !== handler);
    };
  }

  subscribeAll(handler: Subscriber) {
    return this.subscribe("*", handler);
  }

  publish(event: Event) {
    event.timestamp = Date.now();

    if (this.subscribers[event.type]) {
      for (const handler of this.subscribers[event.type]) handler(event);
    }

    if (this.subscribers["*"]) {
      for (const handler of this.subscribers["*"]) handler(event);
    }
  }
}

export const Bus = new EventBus();

// ----------------------
// 🛠 Tool Detection + Execution
// ----------------------

interface ToolCall {
  command: string;
  args: string[];
}

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

async function executeTool(tool: ToolCall): Promise<string> {
  try {
    const allowedCommands = ["ls", "pwd", "df", "uname"];
    if (!allowedCommands.includes(tool.command)) {
      return `Command '${tool.command}' is not allowed`;
    }

    let result: string;
    switch (tool.command) {
      case "ls":
        result = await $`ls ${tool.args}`.text();
        break;
      case "pwd":
        result = await $`pwd`.text();
        break;
      case "df":
        result = await $`df ${tool.args}`.text();
        break;
      case "uname":
        result = await $`uname ${tool.args}`.text();
        break;
      default:
        result = "Command not implemented";
    }
    return result.trim();
  } catch (error: any) {
    return `Error: ${error.message}`;
  }
}

// ----------------------
// 🤖 AI-like Fake Responses
// ----------------------

const responses = [
  "That's an interesting question! Let me think about that...",
  "I understand what you're asking. Here's my perspective:",
  "Based on what you've told me, I would suggest:",
  "That's a great point! To expand on that idea:",
  "Let me help you with that problem step by step:",
];

function generateResponse(message: string): string {
  const randomResponse = responses[Math.floor(Math.random() * responses.length)];
  return `${randomResponse} You said: "${message}". This is a streaming response that demonstrates real-time character-by-character delivery similar to OpenCode's architecture.`;
}

function analyzeToolOutput(command: string, output: string): string {
  switch (command) {
    case "ls": {
      const fileCount = output.split("\n").length - 1;
      return `I can see ${fileCount} items in the current directory.`;
    }
    case "pwd":
      return `You're currently in the ${output} directory.`;
    case "df":
      return "Here's your disk usage information.";
    default:
      return "The command has been executed successfully.";
  }
}

// ----------------------
// 🌐 HTTP Server
// ----------------------

const server = Bun.serve({
  port: 3000,

  async fetch(req: Request): Promise<Response> {
    const url = new URL(req.url);

    // CORS headers
    const corsHeaders = {
      "Access-Control-Allow-Origin": "*",
      "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
      "Access-Control-Allow-Headers": "Content-Type",
    };

    if (req.method === "OPTIONS") {
      return new Response(null, { status: 200, headers: corsHeaders });
    }

    // Health check
    if (url.pathname === "/health" && req.method === "GET") {
      return new Response("Server is running!", { status: 200, headers: corsHeaders });
    }

    // ----------------------
    // CHAT Endpoint
    // ----------------------
    if (url.pathname === "/chat" && req.method === "POST") {
      try {
        const { message } = await req.json();

        // The HTTP stream subscribes to Bus
        const stream = new ReadableStream({
          async start(controller) {
            const encoder = new TextEncoder();

            // Subscribe to all events and forward to client
            const unsubscribe = Bus.subscribeAll((event) => {
              // You can choose to send raw JSON events or plain text per event
              controller.enqueue(encoder.encode(JSON.stringify(event) + "\n"));
            });

            // --- Logic starts ---
            const tool = detectToolCall(message);

            if (tool) {
              Bus.publish({ type: "system.info", data: `Executing ${tool.command}...` });
              Bus.publish({ type: "tool.start", data: { command: tool.command } });

              const output = await executeTool(tool);

              Bus.publish({ type: "tool.output", data: output });
              Bus.publish({ type: "tool.end", data: { command: tool.command } });

              // "AI-like" analysis of tool output
              const analysis = analyzeToolOutput(tool.command, output);
              for (const char of `\n\nBased on tool output: ${analysis}`) {
                Bus.publish({ type: "chat.char", data: char });
                await new Promise((r) => setTimeout(r, 20));
              }
            } else {
              // Normal response (simulated AI streaming)
              const response = generateResponse(message);
              for (const char of response) {
                Bus.publish({ type: "chat.char", data: char });
                await new Promise((r) => setTimeout(r, 30));
              }
            }

            Bus.publish({ type: "chat.complete" });

            unsubscribe();
            controller.close();
          },
        });

        return new Response(stream, {
          headers: {
            "Content-Type": "text/event-stream",
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
            "X-Content-Type-Options": "nosniff",
          },
        });
      } catch (error) {
        console.error("❌ Error processing chat request:", error);
        return new Response("Internal server error", { status: 500, headers: corsHeaders });
      }
    }

    return new Response("Not Found", { status: 404, headers: corsHeaders });
  },
});

// ----------------------
// 🔌 Extra Subscribers (examples)
// ----------------------

// Log all events
Bus.subscribeAll((event) => {
  console.log("📨 Event:", event.type, event.data);
});

// Example: Persist chat messages
Bus.subscribe("chat.char", (event) => {
  // save to DB (demo: console.log)
  // appendMessageToDB(event.data)
});

console.log(`🚀 Chat server running on http://localhost:${server.port}`);
console.log("📡 Endpoints available:");
console.log("   GET  /health - Health check");
console.log("   POST /chat   - Send chat message (streams Bus events)");
console.log("💡 Architecture: EventBus-powered backend (OpenCode-style)");