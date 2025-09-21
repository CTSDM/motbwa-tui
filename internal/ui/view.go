package ui

import (
	"fmt"
	"strings"
)

func (m model) View() string {
	s := "Chat application 못봐\n"
	switch m.flow {
	case initView:
		// here the user select to either login or create user
		// this will be a simple list!
		return "\n" + m.initList.View()

	case signUpView, loginView:
		var b strings.Builder

		for i := range m.credentials {
			b.WriteString(m.credentials[i].View())
			b.WriteRune('\n')
		}

		if m.loginError != "" {
			b.WriteString(m.loginError)
			b.WriteRune('\n')
		}

		return b.String()

	case chatView:
		return fmt.Sprintf(
			"%s\n%s%s%s",
			s,
			m.viewport.View(),
			gap,
			m.textarea.View(),
		)
	}
	return "something went wrong..."
}
