package ui

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/CTSDM/motbwa-tui/internal/ws"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const gap = "\n\n"

type WebSocketMessageReceived struct {
	Message ws.Message
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.flow {
	case initView:
		cmd := m.updateInitView(msg)
		return m, cmd

	case loginView, signUpView:
		cmd := m.updateInputs(msg)
		if m.flow == chatView {
			// websocket connection
			clientManager, err := ws.CreateConnection(m.state)
			if err != nil {
				log.Fatalf("Could not establish the websocket connection: %s", err)
			}
			m.client = clientManager
		}
		return m, cmd

	case addContactView:
		cmd := m.updateContact(msg)
		return m, cmd

	case chatView:
		var (
			tiChatCmd tea.Cmd
			vpChatCmd tea.Cmd
		)

		m.textarea, tiChatCmd = m.textarea.Update(msg)
		m.viewport, vpChatCmd = m.viewport.Update(msg)

		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.viewport.Width = msg.Width
			m.textarea.SetWidth(msg.Width)
			m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(gap)

			if len(m.messages) > 0 {
				// Wrap content before setting it.
				strArray := []string{}
				for _, message := range m.messages {
					strArray = append(strArray, fmt.Sprintf("%s: %s", message.Sender, message.Content))
				}
				m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(strArray, "\n")))
			}
			m.viewport.GotoBottom()

		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyCtrlA:
				m.flow = addContactView
				return m, nil

			case tea.KeyCtrlC:
				m.client.Close()
				fmt.Println(m.textarea.Value())
				return m, tea.Quit

			case tea.KeyEnter:
				msg := m.textarea.Value()
				m.client.SetEgress(msg)
				newMessage := ws.Message{
					Sender:  m.senderStyle.Render("You"),
					Content: m.textarea.Value(),
				}
				m.messages = append(m.messages, newMessage)
				strArray := []string{}
				for _, message := range m.messages {
					strArray = append(strArray, fmt.Sprintf("%s: %s", m.senderStyle.Render(message.Sender), message.Content))
				}
				m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(strArray, "\n")))
				m.textarea.Reset()
				m.viewport.GotoBottom()
			}

		case WebSocketMessageReceived:
			newMessage := ws.Message{
				Sender:  m.senderStyle.Render(msg.Message.Sender),
				Content: msg.Message.Content,
			}
			m.messages = append(m.messages, newMessage)
			strArray := []string{}
			for _, message := range m.messages {
				strArray = append(strArray, fmt.Sprintf("%s: %s", message.Sender, message.Content))
			}
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(strArray, "\n")))
			m.textarea.Reset()
			m.viewport.GotoBottom()
		}

		return m, tea.Batch(tiChatCmd, vpChatCmd, listenToWebSocketMessages(m.client.MessageChannel()))
	}

	return m, nil
}

func (m *model) updateInitView(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.initList.SetWidth(msg.Width)
		return nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return tea.Quit
		case tea.KeyEnter:
			i, ok := m.initList.SelectedItem().(item)
			if ok {
				m.choice = string(i)
				m.flow = m.assignation[m.choice]
			}
			m.initList.Select(0)
			return nil
		}
	}

	var cmd tea.Cmd
	m.initList, cmd = m.initList.Update(msg)
	return cmd
}

func (m *model) updateContact(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.loginError = ""
		switch msg.Type {
		case tea.KeyEnter:
			contactName := m.newContact.Value()
			// user cannot add its own username to the contactlist
			// user can only add other user only once
			if contactName == m.state.User.Username {
				m.loginError = "Cannot add yourself to your contact list"
				return nil
			}
			if _, ok := m.contacts[contactName]; ok {
				m.loginError = "User already in contact list"
				return nil
			}
			if err := m.state.HandlerCheckUserExists(context.Background(), contactName); err != nil {
				m.loginError = err.Error()
				return nil
			}
			m.contacts[contactName] = struct{}{}
			m.newContact.Reset()
			m.flow = chatView
			return nil
		case tea.KeyCtrlB:
			m.flow = chatView
			return nil
		}
	}

	var cmd tea.Cmd
	m.newContact, cmd = m.newContact.Update(msg)
	return cmd
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.credentials))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab, tea.KeyShiftTab:
			// remove focus from previous index
			// with delta and sign we can jump forward/backward
			delta := 1
			sign := 1
			if msg.Type == tea.KeyShiftTab {
				sign = len(m.credentials) - 1
			}
			m.credentials[m.focusIndex].Blur()
			m.focusIndex = (m.focusIndex + delta*sign) % len(m.credentials) * sign
			cmds[m.focusIndex] = m.credentials[m.focusIndex].Focus()
			return tea.Batch(cmds...)

		case tea.KeyEnter:
			m.loginError = ""
			// the enter key validates and then submits the inputs
			username := m.credentials[0].Value()
			password := m.credentials[1].Value()
			// we call the api handler
			switch m.flow {
			case signUpView:
				if err := m.state.HandlerCreateUser(context.Background(), username, password); err != nil {
					m.loginError = err.Error()
					return tea.Batch(cmds...)
				}
			case loginView:
				if err := m.state.HandlerLogin(context.Background(), username, password); err != nil {
					m.loginError = err.Error()
					return tea.Batch(cmds...)
				}
			}

			m.flow += 1
			return m.resetInputs()

		case tea.KeyCtrlC:
			return tea.Quit
		}
	}

	for i := range m.credentials {
		m.credentials[i], cmds[i] = m.credentials[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m *model) resetInputs() tea.Cmd {
	m.focusIndex = 0
	m.loginError = ""
	cmds := make([]tea.Cmd, len(m.credentials))
	for i := range m.credentials {
		m.credentials[i].Reset()
	}
	m.credentials[1].Blur()
	cmds[0] = m.credentials[0].Focus()

	return tea.Batch(cmds...)
}

func listenToWebSocketMessages(msg <-chan ws.Message) tea.Cmd {
	return func() tea.Msg {
		msgChan := <-msg
		return WebSocketMessageReceived{Message: msgChan}
	}
}
