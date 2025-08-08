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
        const body = await req.json() as ChatRequest;
        const { message } = body;

        if (!message || typeof message !== 'string') {
          return new Response('Invalid message', { 
            status: 400, 
            headers: corsHeaders 
          });
        }

        console.log(`📨 Received message: "${message}"`);

        // Generate response text
        const responseText = generateResponse(message);

        // Create streaming response (character by character)
        const stream = new ReadableStream({
          async start(controller) {
            console.log('🚀 Starting stream...');
            
            // Stream each character with a delay to simulate real-time AI response
            for (let i = 0; i < responseText.length; i++) {
              const char = responseText[i];
              controller.enqueue(new TextEncoder().encode(char));
              
              // 50ms delay between characters (configurable)
              await new Promise(resolve => setTimeout(resolve, 50));
            }
            
            console.log('✅ Stream completed');
            controller.close();
          }
        });

        return new Response(stream, {
          status: 200,
          headers: {
            ...corsHeaders,
            'Content-Type': 'text/plain; charset=utf-8',
            'Cache-Control': 'no-cache',
            'Connection': 'keep-alive',
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

console.log(`🚀 Chat server running on http://localhost:${server.port}`);
console.log(`📡 Endpoints available:`);
console.log(`   GET  /health - Health check`);
console.log(`   POST /chat   - Send chat message (streams response)`);
console.log(`💡 Architecture: Similar to OpenCode's backend with streaming responses`);