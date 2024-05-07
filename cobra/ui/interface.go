package ui

import "github.com/spf13/cobra"

type Question struct {
	Text         string
	Options      []string
	Handler      func(string) error
	PasswordType bool
	FilePath     bool
}
type UI interface {
	SetQuestions(questions []Question)
	RunInteractiveUI()
	SetCobra(cobra *cobra.Command)
	AfterPreRun()
	BeforePreRun()
	AfterRun()
	BeforeRun()
	SetLastMessage(msg string)
}

func New() UI {
	return &ui{
		cobra:     nil,
		questions: nil,
		index:     0,
		input:     "",
		cursor:    0,
		errorMsg:  "",
	}
}