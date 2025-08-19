package main

import (
	"bufio"
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

// ===============
// New Event type (matches backend Bus JSON)
// ===============
// type Event struct {
// 	Type      string      `json:"type"`
// 	Data      interface{} `json:"data"`
// 	Timestamp int64       `json:"timestamp"`
// }

// Main application model
type model struct {
	textInput       textinput.Model
	messages        []string
	streaming       bool
	currentResponse string
	width           int
	height          int
	err             error
	program         *tea.Program
	bus             *Bus
}

// Styles using Lipgloss
var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).Padding(0, 1)

	userMessageStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#04B575")).
		Padding(1).MarginBottom(1)

	botMessageStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF6B6B")).
		Padding(1).MarginBottom(1)

	streamingMessageStyle = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#FFD23F")).
		Padding(1).MarginBottom(1)

	inputStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).Italic(true)
)

// Global program reference
var globalProgram *tea.Program

// Initial model
func initialModel(bus *Bus) model {
	ti := textinput.New()
	ti.Placeholder = "Type your message here..."
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 50

	return model{
		textInput: ti,
		messages:  []string{},
		streaming: false,
		bus:       bus,
	}
}

// Init: subscribe to bus and forward all events into Bubble Tea
func (m model) Init() tea.Cmd {
	ch := m.bus.Subscribe()
	go func() {
		for evt := range ch {
			// Send Event directly into Bubble Tea
			globalProgram.Send(evt)
		}
	}()
	return tea.Batch(textinput.Blink, tea.EnterAltScreen)
}

// Update: handle Bubble Tea messages and Events
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = v.Width, v.Height
		return m, nil

	case tea.KeyMsg:
		switch v.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if !m.streaming && strings.TrimSpace(m.textInput.Value()) != "" {
				message := strings.TrimSpace(m.textInput.Value())
				m.messages = append(m.messages, "You: "+message)
				m.textInput.SetValue("")
				m.streaming = true
				return m, sendMessage(message)
			}
		}

	case Event: // <-- Our Bus JSON events
		switch v.Type {
		case "chat.char":
			if s, ok := v.Data.(string); ok {
				m.currentResponse += s
			}
		case "chat.complete":
			if len(m.currentResponse) > 0 {
				m.messages = append(m.messages, "Bot: "+m.currentResponse)
			}
			m.currentResponse = ""
			m.streaming = false
		case "tool.start":
			if data, ok := v.Data.(map[string]interface{}); ok {
				m.currentResponse += fmt.Sprintf("\n\n🔧 Executing: %s\n", data["command"])
			}
		case "tool.output":
			if s, ok := v.Data.(string); ok {
				m.currentResponse += fmt.Sprintf("```\n%s\n```\n", s)
			}
		case "tool.end":
			m.currentResponse += "\n"
		case "stream_err":
			if s, ok := v.Data.(string); ok {
				m.messages = append(m.messages, "Error: "+s)
			}
			m.currentResponse = ""
			m.streaming = false
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// View
func (m model) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("💬 Chat POC - JSON Events") + "\n\n")

	// Show messages
	b.WriteString(m.renderMessages())

	// Current streaming buffer
	if m.streaming {
		streamingText := "Bot: " + m.currentResponse + "▎"
		streaming := streamingMessageStyle.Width(m.width - 4).Render(streamingText)
		b.WriteString(streaming + "\n")
	}

	// Input
	b.WriteString("\n")
	input := inputStyle.Width(m.width - 4).Render(m.textInput.View())
	b.WriteString(input + "\n\n")

	// Help line
	b.WriteString(helpStyle.Render("Press Enter to send • Ctrl+C to quit"))
	return b.String()
}

func (m model) renderMessages() string {
	var b strings.Builder
	start := 0
	maxMessages := 8
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
			b.WriteString(userMessageStyle.Width(width).Render(msg) + "\n")
		} else if strings.HasPrefix(msg, "Bot:") {
			b.WriteString(botMessageStyle.Width(width).Render(msg) + "\n")
		} else {
			b.WriteString(msg + "\n")
		}
	}
	return b.String()
}

// ===============
// sendMessage now decodes JSON per line (backend sends JSON.stringify(event)+"\n")
// ===============
func sendMessage(message string) tea.Cmd {
	return func() tea.Msg {
		reqBody := map[string]string{"message": message}
		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			return Event{Type: "stream_err", Data: "Failed to encode message"}
		}

		req, err := http.NewRequest("POST", "http://localhost:3000/chat", bytes.NewBuffer(jsonBody))
		if err != nil {
			return Event{Type: "stream_err", Data: "Failed to create request"}
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return Event{Type: "stream_err", Data: "Failed to connect to server"}
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return Event{Type: "stream_err", Data: fmt.Sprintf("Server error: %s", resp.Status)}
		}

		go func() {
			defer resp.Body.Close()
			reader := bufio.NewReader(resp.Body)
			for {
				line, err := reader.ReadBytes('\n')
				if err != nil {
					if err == io.EOF {
						appBus.Publish(Event{Type: "stream_end"})
						return
					}
					appBus.Publish(Event{Type: "stream_err", Data: err.Error()})
					return
				}

				// decode JSON line
				var evt Event
				if err := json.Unmarshal(line, &evt); err != nil {
					appBus.Publish(Event{Type: "stream_err", Data: "Invalid JSON: " + err.Error()})
					continue
				}

				// push into Bus
				appBus.Publish(evt)
			}
		}()
		return nil
	}
}