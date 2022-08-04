package smtp

import (
	"bufio"
	"bytes"
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
)

type session struct {
	conn net.Conn

	tls bool
}

func newSession(conn net.Conn) *session {
	return &session{
		conn: conn,
	}
}

func HandleIncoming(conn net.Conn) {
	s := newSession(conn)
	s.Greet()

	for {
		err := s.ReceiveCommand()
		if err != nil {
			log.Error(err)
		}
	}
}

func (s *session) Greet() {
	_, err := s.conn.Write([]byte("220 localhost ESMTP Max\r\n"))
	if err != nil {
		log.Fatal(err)
	}

}

func (s *session) ReceiveCommand() error {
	// buf := make([]byte, command.MaxLen)

	s.ReadLine()
	// n, err := s.conn.Read(buf)
	// if err != nil {
	// 	return err
	// }

	// fmt.Println(string(buf[0:n]))
	return nil
}

func (s *session) ReadLine() {
	sc := bufio.NewScanner(s.conn)
	sc.Split(SplitCRLF)
	for {
		if ok := sc.Scan(); !ok {
			break
		}
		fmt.Println(sc.Bytes())
	}
}

// TODO: Zelf maken
func SplitCRLF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, []byte{'\r', '\n'}); i >= 0 {
		// We have a full newline-terminated line.
		return i + 2, data, nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
