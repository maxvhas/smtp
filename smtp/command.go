package smtp

import (
	"bytes"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
)

const CommandSize = 512

type CommandCode int

const (
	EMPTY CommandCode = iota

	HELO
	EHLO

	MAIL
	RCPT
	DATA

	QUIT
)

var CommandToByte = map[CommandCode][]byte{
	HELO: []byte("HELO"),
	EHLO: []byte("EHLO"),

	MAIL: []byte("MAIL"),
	RCPT: []byte("RCPT"),
	DATA: []byte("DATA"),

	QUIT: []byte("QUIT"),
}

var CommandCodeMap = map[string]CommandCode{
	"HELO": HELO,
	"EHLO": EHLO,
	"MAIL": MAIL,
	"RCPT": RCPT,
	"DATA": DATA,
	"QUIT": QUIT,
}

type Command struct {
	code CommandCode
	word []byte
	args []byte
}

func (c *Command) String() string {
	return fmt.Sprintf("word: %s, args: %s", string(c.word), string(c.args))
}

var (
	ErrCommandTooShort = errors.New("smtp.Parser: Line is too short")
	ErrNoCommandFound  = errors.New("smtp.Parser: No valid command found in line")
)

func ParseCommand(line []byte) (command *Command, err error) {
	word, args, found := bytes.Cut(line, []byte{' '})
	if !found {
		word = line // Handle DATA case
	}

	if code, ok := CommandCodeMap[string(word)]; ok {
		c := &Command{
			code,
			word,
			args,
		}

		log.Info("Command = ", c)

		return c, nil
	}

	return nil, ErrNoCommandFound
}
