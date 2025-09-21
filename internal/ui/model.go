package ui

import (
	"github.com/CTSDM/motbwa-tui/internal/api"
	"github.com/CTSDM/motbwa-tui/internal/ws"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

type flowState int

const (
	signUpView flowState = iota
	initView
	loginView
	chatView
	addContactView
)

type model struct {
	flow  flowState
	state api.State

	// initView
	initList    list.Model
	choice      string
	assignation map[string]flowState

	// contacts
	newContact textinput.Model
	contacts   map[string]struct{}

	// login and create user components
	focusIndex  int
	credentials []textinput.Model
	loginError  string

	// chat components
	client      *ws.ClientManager
	messages    []ws.Message
	textarea    textarea.Model
	viewport    viewport.Model
	senderStyle lipgloss.Style
	err         error
}

func initializeChatView() (textarea.Model, viewport.Model) {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 200

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line sytling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)
	vp.SetContent(`Welcome to the chat room!
Type a message and press Enter to send.`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return ta, vp
}

func initializeAddContactView() textinput.Model {
	t := textinput.New()
	t.Width = 32
	t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	t.Placeholder = "Username"
	t.CharLimit = 32
	t.Focus()
	t.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	t.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return t
}

func initializeInitView(items []string) list.Model {
	// assign the view state
	itemsList := []list.Item{}
	for _, it := range items {
		itemsList = append(itemsList, item(it))
	}

	defaultWidth := 20
	listHeight := 14
	l := list.New(itemsList, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Please select an option"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return l
}

func initializeCredentialsView() []textinput.Model {
	inputs := make([]textinput.Model, 2)

	for i := range inputs {
		t := textinput.New()
		t.Width = 32
		t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		switch i {
		case 0:
			t.Placeholder = "Username"
			t.CharLimit = 32
			t.Focus()
			t.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			t.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

		case 1:
			t.Placeholder = "Password"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '*'
		}

		inputs[i] = t
	}

	return inputs
}

func InitialModel(state api.State) model {
	items := []string{"Login", "Sign up"}
	assignation := map[string]flowState{}
	for i := range items {
		caseLogin := "Login"
		caseSignUp := "Sign up"
		switch items[i] {
		case caseLogin:
			assignation[caseLogin] = loginView
		case caseSignUp:
			assignation[caseSignUp] = signUpView
		}
	}

	taChat, vpChat := initializeChatView()
	tiCredentials := initializeCredentialsView()
	tiContact := initializeAddContactView()
	initList := initializeInitView(items)

	return model{
		// initView
		initList: initList,
		state:    state,
		flow:     initView,
		// login and create user
		credentials: tiCredentials,
		assignation: assignation,

		// new contact
		newContact: tiContact,
		contacts:   make(map[string]struct{}),

		//chat
		textarea:    taChat,
		viewport:    vpChat,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:         nil,
		messages:    make([]ws.Message, 0),
	}
}
