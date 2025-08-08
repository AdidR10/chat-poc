package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Message types for the Bubble Tea update cycle
type (
	streamCharMsg string      // Individual character received from stream
	streamEndMsg  struct{}    // Stream completed
	streamErrMsg  string      // Stream error occurred
)

// Main application model - similar to OpenCode's TUI state management
type model struct {
	textInput       textinput.Model    // User input component
	messages        []string           // Chat history
	streaming       bool               // Is currently receiving streamed response
	currentResponse string             // Buffer for building current response (changed from strings.Builder)
	width           int                // Terminal width
	height          int                // Terminal height
	err             error              // Last error
	program         *tea.Program       // Reference to the program for sending messages
}

// Styles using Lipgloss - similar to OpenCode's terminal styling
var (
	// Title bar style
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	// User message style (similar to OpenCode's user input styling)
	userMessageStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#04B575")).
		Padding(1).
		MarginBottom(1)

	// Bot message style (similar to OpenCode's AI response styling)
	botMessageStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF6B6B")).
		Padding(1).
		MarginBottom(1)

	// Streaming message style (special styling for real-time responses)
	streamingMessageStyle = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#FFD23F")).
		Padding(1).
		MarginBottom(1)

	// Input box style
	inputStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(0, 1)

	// Help text style
	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Italic(true)
)

// Global program reference for streaming
var globalProgram *tea.Program

// Initialize the model - sets up the initial state
func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Type your message here..."
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 50

	return model{
		textInput: ti,
		messages:  []string{},
		streaming: false,
	}
}

// Init command - runs when the program starts
func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, tea.EnterAltScreen)
}

// Update handles all events - this is the core of Bubble Tea's architecture
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Handle terminal resize
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			// Quit the application
			return m, tea.Quit

		case tea.KeyEnter:
			// Send message if not currently streaming and input is not empty
			if !m.streaming && strings.TrimSpace(m.textInput.Value()) != "" {
				message := strings.TrimSpace(m.textInput.Value())
				
				// Add user message to history
				m.messages = append(m.messages, "You: "+message)
				
				// Clear input and set streaming state
				m.textInput.SetValue("")
				m.streaming = true
				
				// Send message to backend
				return m, sendMessage(message)
			}
		}

	case streamCharMsg:
		// Received a character from the stream
		m.currentResponse += string(msg)  // Changed from WriteString to +=
		return m, nil

	case streamEndMsg:
		// Stream completed successfully
		if len(m.currentResponse) > 0 {  // Changed from m.currentResponse.Len()
			m.messages = append(m.messages, "Bot: "+m.currentResponse)  // Changed from String()
		}
		m.currentResponse = ""  // Changed from Reset()
		m.streaming = false
		return m, nil

	case streamErrMsg:
		// Stream error occurred
		m.messages = append(m.messages, "Error: "+string(msg))
		m.currentResponse = ""  // Changed from Reset()
		m.streaming = false
		return m, nil
	}

	// Update text input
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// View renders the UI - similar to OpenCode's TUI rendering
func (m model) View() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("💬 Chat POC - OpenCode Architecture Demo")
	b.WriteString(title + "\n\n")

	// Calculate available space for messages
	contentHeight := m.height - 8 // Reserve space for title, input, and help
	if contentHeight < 1 {
		contentHeight = 10
	}

	// Message display area
	messageArea := m.renderMessages()
	b.WriteString(messageArea)

	// Current streaming response (if any)
	if m.streaming {
		streamingText := "Bot: " + m.currentResponse + "▎" // Changed from String()
		streaming := streamingMessageStyle.Width(m.width - 4).Render(streamingText)
		b.WriteString(streaming + "\n")
	}

	// Input area
	b.WriteString("\n")
	input := inputStyle.Width(m.width - 4).Render(m.textInput.View())
	b.WriteString(input + "\n\n")

	// Help text
	help := helpStyle.Render("Press Enter to send • Ctrl+C to quit • Characters stream in real-time")
	b.WriteString(help)

	return b.String()
}

// renderMessages displays the chat history with proper styling
func (m model) renderMessages() string {
	var b strings.Builder
	
	// Show the last few messages to fit in the terminal
	start := 0
	maxMessages := 8 // Show last 8 messages
	if len(m.messages) > maxMessages {
		start = len(m.messages) - maxMessages
	}

	for i := start; i < len(m.messages); i++ {
		msg := m.messages[i]
		width := m.width - 4
		if width < 20 {
			width = 20
		}

		if strings.HasPrefix(msg, "You:") {
			styled := userMessageStyle.Width(width).Render(msg)
			b.WriteString(styled + "\n")
		} else if strings.HasPrefix(msg, "Bot:") {
			styled := botMessageStyle.Width(width).Render(msg)
			b.WriteString(styled + "\n")
		} else {
			// Error messages or other types
			b.WriteString(msg + "\n")
		}
	}

	return b.String()
}

// sendMessage sends a message to the backend and starts streaming response
func sendMessage(message string) tea.Cmd {
	return func() tea.Msg {
		// Prepare request payload
		requestBody := map[string]string{
			"message": message,
		}
		
		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			return streamErrMsg("Failed to encode message")
		}

		// Create HTTP request WITHOUT context timeout for streaming
		req, err := http.NewRequest("POST", "http://localhost:3000/chat", bytes.NewBuffer(jsonBody))
		if err != nil {
			return streamErrMsg("Failed to create request")
		}
		req.Header.Set("Content-Type", "application/json")

		// Send the request
		client := &http.Client{
			Timeout: 60 * time.Second, // Set timeout on client instead
		}
		resp, err := client.Do(req)
		if err != nil {
			return streamErrMsg("Failed to connect to server. Make sure backend is running on localhost:3000!")
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return streamErrMsg(fmt.Sprintf("Server error: %s", resp.Status))
		}

		// Start streaming
		go func() {
			defer resp.Body.Close()
			
			// Read response in small chunks to simulate streaming
			buf := make([]byte, 1)
			for {
				n, err := resp.Body.Read(buf)
				if err != nil {
					if err == io.EOF {
						if globalProgram != nil {
							globalProgram.Send(streamEndMsg{})
						}
						return
					}
					if globalProgram != nil {
						globalProgram.Send(streamErrMsg("Stream read error: " + err.Error()))
					}
					return
				}
				
				if n > 0 && globalProgram != nil {
					globalProgram.Send(streamCharMsg(string(buf[:n])))
					time.Sleep(time.Millisecond * 50) // Slightly longer delay to see streaming effect
				}
			}
		}()

		return nil
	}
}