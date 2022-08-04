package command

import (
	"errors"
	"strings"
)

const CommandSize = 512

type CommandWord string

const (
	HELO CommandWord = "HELO"
	EHLO             = "EHLO"

	FROM = ""
	TO
	DATA

	QUIT
)

type Command struct {
	word string
	args []string
}

var ErrCommandTooShort = errors.New("Parse error")

func ParseCommand(message string) (*Command, error) {
	split := strings.Split(message, " ")

	if len(split) < 2 {
		return nil, ErrCommandTooShort
	}

	return &Command{
		split[0],
		split[1:],
	}, nil
}
