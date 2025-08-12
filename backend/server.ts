/**
 * Simple chat server with streaming responses
 * Mirrors OpenCode's backend architecture with Bun runtime
 */

interface ChatRequest {
  message: string;
}

interface ChatResponse {
  reply: string;
}

// Tool call interface
interface ToolCall {
  command: string;
  args: string[];
}

// Simple pattern matching for tool triggers
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

// Execute tool safely
import { $ } from "bun";

// Safe command execution with whitelisted commands
async function executeTool(tool: ToolCall): Promise<string> {
  try {
    // Whitelist of allowed commands for safety
    const allowedCommands = ["ls", "pwd", "df", "uname"];
    
    if (!allowedCommands.includes(tool.command)) {
      return `Command '${tool.command}' is not allowed`;
    }
    
    // Execute based on command
    let result: string;
    
    switch(tool.command) {
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

// Simple AI-like responses for demonstration
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

const server = Bun.serve({
  port: 3000,
  
  async fetch(req: Request): Promise<Response> {
    const url = new URL(req.url);
    
    // CORS headers for cross-origin requests
    const corsHeaders = {
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
      'Access-Control-Allow-Headers': 'Content-Type',
    };

    // Handle CORS preflight requests
    if (req.method === 'OPTIONS') {
      return new Response(null, { 
        status: 200, 
        headers: corsHeaders 
      });
    }

    // Health check endpoint
    if (url.pathname === '/health' && req.method === 'GET') {
      return new Response('Server is running!', { 
        status: 200, 
        headers: corsHeaders 
      });
    }

    // Chat endpoint with streaming response
      if (url.pathname === '/chat' && req.method === 'POST') {
      try {
        const { message } = await req.json();
      
        // Create a ReadableStream for response
        const stream = new ReadableStream({
          async start(controller) {
            const encoder = new TextEncoder();
            
            // Check if message requires a tool
            const tool = detectToolCall(message);
            
            if (tool) {
              // Stream pre-tool message
              const preMessage = `I'll execute the ${tool.command} command for you...\n\n`;
              for (const char of preMessage) {
                controller.enqueue(encoder.encode(char));
                await new Promise(resolve => setTimeout(resolve, 20));
              }
              
              // Send tool call start marker
              controller.enqueue(encoder.encode(`[TOOL_START:${tool.command}]`));
              await new Promise(resolve => setTimeout(resolve, 50));
              
              // Execute tool
              const output = await executeTool(tool);
              
              // Send tool output
              controller.enqueue(encoder.encode(`[TOOL_OUTPUT:${output}]`));
              await new Promise(resolve => setTimeout(resolve, 50));
              
              // Send tool end marker
              controller.enqueue(encoder.encode("[TOOL_END]"));
              await new Promise(resolve => setTimeout(resolve, 50));
              
              // Stream post-tool analysis
              const analysis = `\n\nBased on the output, ${analyzeToolOutput(tool.command, output)}`;
              for (const char of analysis) {
                controller.enqueue(encoder.encode(char));
                await new Promise(resolve => setTimeout(resolve, 20));
              }
            } else {
              // Normal response for non-tool messages
              const response = generateResponse(message);
              for (const char of response) {
                controller.enqueue(encoder.encode(char));
                await new Promise(resolve => setTimeout(resolve, 30));
              }
            }
            
            controller.close();
          }
        });

        return new Response(stream, {
          headers: { 
            "Content-Type": "text/plain",
            "X-Content-Type-Options": "nosniff"
          }
        });

      } catch (error) {
        console.error('❌ Error processing chat request:', error);
        return new Response('Internal server error', { 
          status: 500, 
          headers: corsHeaders 
        });
      }
    }

    // Default 404 response
    return new Response('Not Found', { 
      status: 404, 
      headers: corsHeaders 
    });
  }
});


// Simple analysis function
function analyzeToolOutput(command: string, output: string): string {
  switch(command) {
    case "ls":
      const fileCount = output.split('\n').length - 1;
      return `I can see ${fileCount} items in the current directory.`;
    case "pwd":
      return `you're currently in the ${output} directory.`;
    case "df":
      return "here's your disk usage information.";
    default:
      return "the command has been executed successfully.";
  }
}

console.log(`🚀 Chat server running on http://localhost:${server.port}`);
console.log(`📡 Endpoints available:`);
console.log(`   GET  /health - Health check`);
console.log(`   POST /chat   - Send chat message (streams response)`);
console.log(`💡 Architecture: Similar to OpenCode's backend with streaming responses`);